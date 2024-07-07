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

var growTempLast float64
var growHumidLast float64
var growTempHistory map[time.Time]float64
var growHumidHistory map[time.Time]float64

func (ui *GrowUi) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	switch msg.Topic() {
	case "growroom/room/temp":
		growTemp, err := parseValue(msg)
		if err != nil {
			log.Println("Could not parse MQTT message")
			return
		}
		growTempLast = growTemp
		growTempHistory[time.Now()] = growTemp
	case "growroom/room/humid":
		growHumid, err := parseValue(msg)
		if err != nil {
			log.Println("Could not parse MQTT message")
			return
		}
		growHumidLast = growHumid
		growHumidHistory[time.Now()] = growHumid
	}

	if len(growTempHistory) > 1 && len(growHumidHistory) > 1 {
		ui.renderGraph()
	}
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
	tempHistoryTimes, tempHistoryValues := mapToGraphSlice(growTempHistory)
	humidHistoryTimes, humidHistoryValues := mapToGraphSlice(growHumidHistory)

	graph := chart.Chart{
		Background: chart.Style{FillColor: chart.ColorTransparent},
		DPI:        200,
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
			FillColor: drawing.ColorTransparent,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    "Temperature",
				XValues: tempHistoryTimes,
				YValues: tempHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorRed,
					StrokeWidth: 8,
				},
			},
			chart.TimeSeries{
				Name:    "Relative Humidity",
				XValues: humidHistoryTimes,
				YValues: humidHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorBlue,
					StrokeWidth: 8,
				},
			},
		},
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
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

	growTempHistory = make(map[time.Time]float64)
	growHumidHistory = make(map[time.Time]float64)

	// Connect to mqtt
	broker := "homeserver"
	port := 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("screen-app2")
	opts.SetDefaultPublishHandler(ui.messagePubHandler)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	log.Println("Connected to MQTT")

	client.Subscribe("growroom/room/temp", 1, nil)
	client.Subscribe("growroom/room/humid", 2, nil)
	log.Printf("Subscribed to box temp/humid")
}

func (ui *GrowUi) Bounds() (width, height int) {
	//return WIDTH, fontHeight + linePadding
	return WIDTH, HEIGHT
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
		fmt.Sprintf("%.2f temp %.2f rh", growTempLast, growHumidLast),
		defaultFont,
		0,
		500,
		textColor,
	)

	text.Draw(
		ui.screen,
		fmt.Sprintf("%.2f vpd", calculateVPD(growTempLast, growHumidLast)),
		defaultFont,
		0,
		500+fontHeight+linePadding,
		textColor,
	)

	return ui.screen
}
