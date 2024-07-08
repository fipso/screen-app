package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type GrowUi struct {
	screen *ebiten.Image

	currentGraphImage *ebiten.Image
}

var growRoomTempLast float64
var growRoomHumidLast float64
var growRoomTempHistory map[time.Time]float64
var growRoomHumidHistory map[time.Time]float64

var growBoxTempLast float64
var growBoxHumidLast float64
var growBoxTempHistory map[time.Time]float64
var growBoxHumidHistory map[time.Time]float64

func (ui *GrowUi) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	switch msg.Topic() {
	case "growroom/room/temp":
		growTemp, err := parseValue(msg)
		if err != nil {
			log.Println("Could not parse MQTT message")
			return
		}
		growRoomTempLast = growTemp
		growRoomTempHistory[time.Now()] = growTemp
	case "growroom/room/humid":
		growHumid, err := parseValue(msg)
		if err != nil {
			log.Println("Could not parse MQTT message")
			return
		}
		growRoomHumidLast = growHumid
		growRoomHumidHistory[time.Now()] = growHumid
	case "growbox/sensor/box_temperature/state":
		growTemp, err := parseValue(msg)
		if err != nil {
			log.Println("Could not parse MQTT message")
			return
		}
		growBoxTempLast = growTemp
		growBoxTempHistory[time.Now()] = growTemp
	case "growbox/sensor/box_humidity/state":
		growHumid, err := parseValue(msg)
		if err != nil {
			log.Println("Could not parse MQTT message")
			return
		}
		growBoxHumidLast = growHumid
		growBoxHumidHistory[time.Now()] = growHumid
	}

	ui.renderGraph()
}

func parseValue(msg mqtt.Message) (float64, error) {
	s := fmt.Sprintf("%s", msg.Payload())
	return strconv.ParseFloat(s, 64)
}

func mapToGraphSlice(inputMap map[time.Time]float64) ([]time.Time, []float64) {
	var times []time.Time
	var values []float64

	for k, _ := range inputMap {
		times = append(times, k)
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	for _, t := range times {
		values = append(values, inputMap[t])
	}

	return times, values
}

func (ui *GrowUi) renderGraph() {
	tempRoomHistoryTimes, tempRoomHistoryValues := mapToGraphSlice(growRoomTempHistory)
	humidRoomHistoryTimes, humidRoomHistoryValues := mapToGraphSlice(growRoomHumidHistory)

	tempBoxHistoryTimes, tempBoxHistoryValues := mapToGraphSlice(growBoxTempHistory)
	humidBoxHistoryTimes, humidBoxHistoryValues := mapToGraphSlice(growBoxHumidHistory)

	graph := chart.Chart{
		// change font color to white
		DPI:        150,
		Background: chart.Style{FillColor: chart.ColorTransparent},
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeMinuteValueFormatter,
		},
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 10.0,
				Max: 80.0,
			},
		},
		Canvas: chart.Style{
			FillColor: drawing.Color{R: bgColor.R, G: bgColor.G, B: bgColor.B, A: bgColor.A},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    "Room Temp",
				XValues: tempRoomHistoryTimes,
				YValues: tempRoomHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorRed,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "Room RH",
				XValues: humidRoomHistoryTimes,
				YValues: humidRoomHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorBlue,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "Box Temp",
				XValues: tempBoxHistoryTimes,
				YValues: tempBoxHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorOrange,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "Box RH",
				XValues: humidBoxHistoryTimes,
				YValues: humidBoxHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorAlternateBlue,
					StrokeWidth: 6,
				},
			},
		},
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph, chart.Style{
			FillColor: graph.Background.FillColor,
		}),
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		log.Println("Could not render graph")
		log.Println(err)
	}

	img, err := png.Decode(buffer)
	if err != nil {
		log.Println("Could not decode graph png")
	}

	ui.currentGraphImage = ebiten.NewImageFromImage(img)
}

func calculateVPD(T, RH float64) float64 {
	// Calculate the saturated vapor pressure (es) in kPa
	es := (610.7 * math.Pow(10, (7.5*T)/(237.3+T))) / 1000

	// Calculate the actual vapor pressure (ea) in kPa
	ea := es * (RH / 100)

	// Calculate VPD in kPa
	vpd := es - ea

	return vpd
}

func (ui *GrowUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)

	growRoomTempHistory = make(map[time.Time]float64)
	growRoomHumidHistory = make(map[time.Time]float64)
	growBoxTempHistory = make(map[time.Time]float64)
	growBoxHumidHistory = make(map[time.Time]float64)

	// Connect to mqtt
	broker := "homeserver"
	port := 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(fmt.Sprintf("screen-app-%d", time.Now().Unix()))
	opts.SetDefaultPublishHandler(ui.messagePubHandler)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	log.Println("Connected to MQTT")

	client.Subscribe("growroom/room/temp", 0, nil)
	client.Subscribe("growroom/room/humid", 0, nil)
	client.Subscribe("growbox/sensor/box_temperature/state", 0, nil)
	client.Subscribe("growbox/sensor/box_humidity/state", 0, nil)
	log.Printf("Subscribed to room and box temp/humid")
}

func (ui *GrowUi) Bounds() (width, height int) {
	return WIDTH, 1000
}

func (ui *GrowUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	// Plot the temperature and humidity history
	if ui.currentGraphImage != nil {
		pos := ebiten.GeoM{}
		pos.Translate(0, float64(linePadding*2))
		opts := &ebiten.DrawImageOptions{
			GeoM: pos,
		}
		ui.screen.DrawImage(ui.currentGraphImage, opts)
	}

	text.Draw(
		ui.screen,
		fmt.Sprintf(
			"room\n%.2f temp %.2f rh\n%.2f vpd",
			growRoomTempLast,
			growRoomHumidLast,
			calculateVPD(growRoomTempLast, growRoomHumidLast),
		),
		defaultFont,
		0,
		500,
		textColor,
	)

	text.Draw(
		ui.screen,
		fmt.Sprintf(
			"box\n%.2f temp %.2f rh\n%.2f vpd",
			growBoxTempLast,
			growBoxHumidLast,
			calculateVPD(growBoxTempLast, growBoxHumidLast),
		),
		defaultFont,
		0,
		775,
		textColor,
	)

	return ui.screen
}
