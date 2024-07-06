package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"strconv"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/wcharczuk/go-chart/v2"
)

type GrowUi struct {
	screen *ebiten.Image

	currentGraphImage *ebiten.Image
}

var growTemp float64
var growTempHistory []float64

var growHumid float64
var growHumidHistory []float64

func (ui *GrowUi) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	var err error
	switch msg.Topic() {
	case "growroom/room/temp":
		growTemp, err = parseValue(msg)
		if err != nil {
			log.Println("Could not parse MQTT message")
			return
		}
		growTempHistory = append(growTempHistory, growTemp)
	case "growroom/room/humid":
		growHumid, err = parseValue(msg)
		if err != nil {
			log.Println("Could not parse MQTT message")
			return
		}
		growHumidHistory = append(growHumidHistory, growHumid)
	}

	if len(growTempHistory) > 1 && len(growHumidHistory) > 1 {
		ui.renderGraph()
	}
}

func parseValue(msg mqtt.Message) (float64, error) {
	s := fmt.Sprintf("%s", msg.Payload())
	return strconv.ParseFloat(s, 64)
}

func (ui *GrowUi) renderGraph() {
	var n []float64
	for i := 0; i < len(growTempHistory); i++ {
		n = append(n, float64(i))
	}

	graph := chart.Chart{
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: n,
				YValues: growTempHistory,
			},
			chart.ContinuousSeries{
				XValues: n,
				YValues: growHumidHistory,
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
		fmt.Sprintf("%.2f temp %.2f rh", growTemp, growHumid),
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
