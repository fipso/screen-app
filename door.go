package main

import (
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

type DoorService struct {
	lastRing time.Time
}

func (s *DoorService) Run() {
	mqttService.WaitReady()

	mqttService.Client.Subscribe("door/ring", 0, func(client mqtt.Client, msg mqtt.Message) {
		// Play door alarm sound
		go s.playAlarm()

		defer func() {
			s.lastRing = time.Now()
		}()
		if game == nil || time.Since(s.lastRing) < 21*time.Second {
			return
		}

		// Create new door alert modal
		m := &ModalUi{
			stackLayout: []UiElement{
				&AlertUi{
					msg: "  faggot on the\n     doooooor",
				},
			},
		}
		m.Init()
		game.currentModal = m

		// Remove door alert modal after 20s
		go func() {
			time.Sleep(time.Second * 20)
			game.currentModal = nil
		}()
	})
}

func (s *DoorService) playAlarm() {
	// Play door alarm sound mp3
	soundFileReader, err := os.Open("./assets/alaram.mp3")
	if err != nil {
		log.Println(err)
		return
	}
	stream, err := mp3.DecodeWithSampleRate(44100, soundFileReader)
	if err != nil {
		log.Println(err)
		return
	}
	ac := audio.NewContext(44100)
	player, err := ac.NewPlayer(stream)
	if err != nil {
		log.Println(err)
		return
	}
	player.SetVolume(1)
	player.Play()
}
