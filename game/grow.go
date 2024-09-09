package game

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

	tempGraph      *chart.Chart
	tempGraphImage *ebiten.Image
	vpdGraph       *chart.Chart
	vpdGraphImage  *ebiten.Image
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

	// Redraw the graph
	ui.tempGraph = ui.buildTempGraph()
	ui.vpdGraph = ui.buildVpdGraph()

	ui.tempGraphImage = ui.renderGraph(ui.tempGraph)
	ui.vpdGraphImage = ui.renderGraph(ui.vpdGraph)
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

func (ui *GrowUi) buildTempGraph() *chart.Chart {
	tempRoomHistoryTimes, tempRoomHistoryValues := mapToGraphSlice(growRoomTempHistory)
	humidRoomHistoryTimes, humidRoomHistoryValues := mapToGraphSlice(growRoomHumidHistory)

	tempBoxHistoryTimes, tempBoxHistoryValues := mapToGraphSlice(growBoxTempHistory)
	humidBoxHistoryTimes, humidBoxHistoryValues := mapToGraphSlice(growBoxHumidHistory)

	// Add outdoor RH for room times
	outdoorHumidLine := map[time.Time]float64{}
	if weatherCurrentData != nil && len(tempRoomHistoryTimes) > 0 {
		rh := weatherCurrentData.Weather.RelativeHumidity
		outdoorHumidLine[tempRoomHistoryTimes[0]] = rh
		outdoorHumidLine[tempRoomHistoryTimes[len(tempRoomHistoryTimes)-1]] = rh
	}
	outdoorHumidTimes, outdoorHumidValues := mapToGraphSlice(outdoorHumidLine)

	// Add outdoor Temp for room times
	outdoorTempLine := map[time.Time]float64{}
	if weatherCurrentData != nil && len(tempRoomHistoryTimes) > 0 {
		temp := weatherCurrentData.Weather.Temperature
		outdoorTempLine[tempRoomHistoryTimes[0]] = temp
		outdoorTempLine[tempRoomHistoryTimes[len(tempRoomHistoryTimes)-1]] = temp
	}
	outdoorTempTimes, outdoorTempValues := mapToGraphSlice(outdoorTempLine)

	graph := chart.Chart{
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeMinuteValueFormatter,
		},
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 15.0,
				Max: 90.0,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    "Room Temp",
				XValues: tempRoomHistoryTimes,
				YValues: tempRoomHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorOrange,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "Room RH",
				XValues: humidRoomHistoryTimes,
				YValues: humidRoomHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorOrange,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "Outdoor RH",
				XValues: outdoorHumidTimes,
				YValues: outdoorHumidValues,
				Style: chart.Style{
					StrokeColor: chart.ColorGreen,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "Box Temp",
				XValues: tempBoxHistoryTimes,
				YValues: tempBoxHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorBlue,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "Box RH",
				XValues: humidBoxHistoryTimes,
				YValues: humidBoxHistoryValues,
				Style: chart.Style{
					StrokeColor: chart.ColorBlue,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "Outdoor Temp",
				XValues: outdoorTempTimes,
				YValues: outdoorTempValues,
				Style: chart.Style{
					StrokeColor: chart.ColorGreen,
					StrokeWidth: 6,
				},
			},
		},
	}

	return &graph
}

func (ui *GrowUi) buildVpdGraph() *chart.Chart {
	tempRoomHistoryTimes, tempRoomHistoryValues := mapToGraphSlice(growRoomTempHistory)
	_, humidRoomHistoryValues := mapToGraphSlice(growRoomHumidHistory)

	tempBoxHistoryTimes, tempBoxHistoryValues := mapToGraphSlice(growBoxTempHistory)
	_, humidBoxHistoryValues := mapToGraphSlice(growBoxHumidHistory)

	vpdMinLine := map[time.Time]float64{}
	vpdMaxLine := map[time.Time]float64{}
	if len(tempRoomHistoryTimes) > 0 {
		vpdMinLine[tempRoomHistoryTimes[0]] = 0.4
		vpdMinLine[tempRoomHistoryTimes[len(tempRoomHistoryTimes)-1]] = 0.4

		vpdMaxLine[tempRoomHistoryTimes[0]] = 1.2
		vpdMaxLine[tempRoomHistoryTimes[len(tempRoomHistoryTimes)-1]] = 1.2
	}

	vpdMinTimes, vpdMinValues := mapToGraphSlice(vpdMinLine)
	vpdMaxTimes, vpdMaxValues := mapToGraphSlice(vpdMaxLine)

	vpdRoomLine := map[time.Time]float64{}
	for i, t := range tempRoomHistoryTimes {
		if i >= len(humidRoomHistoryValues) {
			break
		}
		vpdRoomLine[t] = calculateVPD(tempRoomHistoryValues[i], humidRoomHistoryValues[i])
	}

	vpdBoxLine := map[time.Time]float64{}
	for i, t := range tempBoxHistoryTimes {
		if i >= len(humidBoxHistoryValues) {
			break
		}
		vpdBoxLine[t] = calculateVPD(tempBoxHistoryValues[i], humidBoxHistoryValues[i])
	}

	vpdRoomTimes, vpdRoomValues := mapToGraphSlice(vpdRoomLine)
	vpdBoxTimes, vpdBoxValues := mapToGraphSlice(vpdBoxLine)

	graph := chart.Chart{
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeMinuteValueFormatter,
		},
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0.2,
				Max: 1.75,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    "VPD Min",
				XValues: vpdMinTimes,
				YValues: vpdMinValues,
				Style: chart.Style{
					StrokeColor: chart.ColorRed,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "VPD Room",
				XValues: vpdRoomTimes,
				YValues: vpdRoomValues,
				Style: chart.Style{
					StrokeColor: chart.ColorOrange,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "VPD Box",
				XValues: vpdBoxTimes,
				YValues: vpdBoxValues,
				Style: chart.Style{
					StrokeColor: chart.ColorBlue,
					StrokeWidth: 6,
				},
			},
			chart.TimeSeries{
				Name:    "VPD Max",
				XValues: vpdMaxTimes,
				YValues: vpdMaxValues,
				Style: chart.Style{
					StrokeColor: chart.ColorRed,
					StrokeWidth: 6,
				},
			},
		},
	}

	return &graph
}

func (ui *GrowUi) renderGraph(graph *chart.Chart) *ebiten.Image {
	// Apply defaults
	graph.DPI = 150
	graph.Background = chart.Style{FillColor: chart.ColorTransparent}
	graph.Canvas = chart.Style{
		FillColor: drawing.Color{R: bgColor.R, G: bgColor.G, B: bgColor.B, A: bgColor.A},
	}

	graph.Elements = []chart.Renderable{
		chart.Legend(graph, chart.Style{
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

	return ebiten.NewImageFromImage(img)
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

	ui.tempGraph = ui.buildTempGraph()
	ui.tempGraphImage = ui.renderGraph(ui.tempGraph)

	ui.vpdGraph = ui.buildVpdGraph()
	ui.vpdGraphImage = ui.renderGraph(ui.vpdGraph)

	ui.SubscribeGrowMqtt()
}

func (ui *GrowUi) SubscribeGrowMqtt() {
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
	return WIDTH, 1420
}

func (ui *GrowUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	// Plot the temperature and humidity history
	if ui.tempGraphImage != nil {
		pos := ebiten.GeoM{}
		pos.Translate(0, float64(linePadding*2))
		opts := &ebiten.DrawImageOptions{
			GeoM: pos,
		}
		ui.screen.DrawImage(ui.tempGraphImage, opts)
	}

	if ui.vpdGraphImage != nil {
		pos := ebiten.GeoM{}
		pos.Translate(0, 1000)
		opts := &ebiten.DrawImageOptions{
			GeoM: pos,
		}
		ui.screen.DrawImage(ui.vpdGraphImage, opts)
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
		520,
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
		800,
		textColor,
	)

	return ui.screen
}
