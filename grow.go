package main

import (
        "log"
        "fmt"

        "github.com/eclipse/paho.mqtt.golang"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type GrowUi struct {
  screen *ebiten.Image
}

var growTemp string
var growHumid string

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    switch msg.Topic() {
    case "growroom/room/temp":
      growTemp = fmt.Sprintf("%s", msg.Payload())
    case "growroom/room/humid":
      growHumid = fmt.Sprintf("%s", msg.Payload())
    }

}

func (ui *GrowUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)

        // Connect to mqtt
        broker := "homeserver"
        port := 1883
        opts := mqtt.NewClientOptions()
        opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
        opts.SetClientID("screen-app")
        opts.SetDefaultPublishHandler(messagePubHandler)
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
	return WIDTH, fontHeight + linePadding
}

func (ui *GrowUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)
	text.Draw(ui.screen, fmt.Sprintf("%s temp %s rh", growTemp, growHumid), defaultFont, 0, fontHeight, textColor)

	return ui.screen
}
