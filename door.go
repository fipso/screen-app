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
	lastOpen     time.Time
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

	// Use mqttService.On so the subscription is re-established by the
	// OnConnect handler if the broker connection drops.
	mqttService.On("door/ring", func(client mqtt.Client, msg mqtt.Message) {
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

		// Remove door alert modal after 20s — only if it's still ours, so a
		// later modal (e.g. door/open) doesn't get cleared by our timer.
		go func() {
			time.Sleep(time.Second * 20)
			if game != nil && game.currentModal == m {
				game.currentModal = nil
			}
		}()
	})

	mqttService.On("door/open", func(client mqtt.Client, msg mqtt.Message) {
		if game == nil || time.Since(s.lastOpen) < 4*time.Second {
			return
		}
		s.lastOpen = time.Now()

		// Don't stomp on an active doorbell modal.
		if game.currentModal != nil {
			return
		}

		m := &ModalUi{
			stackLayout: []UiElement{
				&AlertUi{
					icon: "", // FA door-open
				},
			},
		}
		m.Init()
		game.currentModal = m

		go func() {
			time.Sleep(3 * time.Second)
			if game != nil && game.currentModal == m {
				game.currentModal = nil
			}
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
