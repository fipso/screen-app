package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/wcharczuk/go-chart/v2"
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
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: tempHistoryTimes,
				YValues: tempHistoryValues,
			},
			chart.TimeSeries{
				XValues: humidHistoryTimes,
				YValues: humidHistoryValues,
			},
		},
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
	text.Draw(
		ui.screen,
		fmt.Sprintf("%.2f temp %.2f rh", growTempLast, growHumidLast),
		defaultFont,
		0,
		fontHeight,
		textColor,
	)

	// Plot the temperature and humidity history
	if ui.currentGraphImage != nil {
		pos := ebiten.GeoM{}
		pos.Translate(0, float64(fontHeight+linePadding))
		opts := &ebiten.DrawImageOptions{
			GeoM: pos,
		}
		ui.screen.DrawImage(ui.currentGraphImage, opts)
	}

	return ui.screen
}
