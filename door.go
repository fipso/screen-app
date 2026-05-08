package main

import (
	"bytes"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

type DoorService struct {
	lastRing     time.Time
	audioContext *audio.Context
	alarmBytes   []byte
}

func (s *DoorService) Run() {
	s.audioContext = audio.NewContext(44100)

	// Preload the alarm mp3 once so playAlarm doesn't open a fd per ring.
	if b, err := os.ReadFile("./assets/alaram.mp3"); err != nil {
		log.Println("door: could not load alarm mp3:", err)
	} else {
		s.alarmBytes = b
	}

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
	if len(s.alarmBytes) == 0 {
		return
	}
	stream, err := mp3.DecodeWithSampleRate(44100, bytes.NewReader(s.alarmBytes))
	if err != nil {
		log.Println(err)
		return
	}
	player, err := s.audioContext.NewPlayer(stream)
	if err != nil {
		log.Println(err)
		return
	}
	player.SetVolume(1)
	player.Play()

	// Tear the player down once playback finishes so the decoder + buffers
	// don't leak on every ring.
	go func() {
		for player.IsPlaying() {
			time.Sleep(200 * time.Millisecond)
		}
		if err := player.Close(); err != nil {
			log.Println("door: player close:", err)
		}
	}()
}
