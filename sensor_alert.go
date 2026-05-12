package main

import (
	"bytes"
	"log"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

type SensorAlertService struct {
	audioContext *audio.Context
	sounds       map[string][]byte
	states       map[string]*sensorAlertState
	byTopic      map[string][]*sensorAlertState
}

type sensorAlertSide string

const (
	alertSideNone sensorAlertSide = ""
	alertSideMin  sensorAlertSide = "min"
	alertSideMax  sensorAlertSide = "max"
)

type sensorAlertState struct {
	cfg         SensorAlertConfig
	triggered   bool
	triggerSide sensorAlertSide
	haveSample  bool
	activeModal *ModalUi
}

func (s *SensorAlertService) Run() {
	s.audioContext = getAudioContext()
	s.sounds = map[string][]byte{}
	s.states = map[string]*sensorAlertState{}
	s.byTopic = map[string][]*sensorAlertState{}

	for _, cfg := range config.Alerts {
		if cfg.Min == nil && cfg.Max == nil {
			log.Printf("alert %s: skipping; both min and max are nil", cfg.Name)
			continue
		}
		if cfg.Sound != "" {
			if _, ok := s.sounds[cfg.Sound]; !ok {
				b, err := os.ReadFile(cfg.Sound)
				if err != nil {
					log.Printf("alert %s: could not load sound %s: %v", cfg.Name, cfg.Sound, err)
				} else {
					s.sounds[cfg.Sound] = b
				}
			}
		}
		st := &sensorAlertState{cfg: cfg}
		s.states[cfg.Name] = st
		s.byTopic[cfg.Topic] = append(s.byTopic[cfg.Topic], st)
	}

	mqttService.WaitReady()

	for topic, sts := range s.byTopic {
		t := topic
		names := make([]string, 0, len(sts))
		for _, st := range sts {
			names = append(names, st.cfg.Name)
		}
		log.Printf("alert: subscribing to %s for %v", t, names)
		mqttService.On(t, func(client mqtt.Client, msg mqtt.Message) {
			s.handleMessage(t, msg.Payload())
		})
	}
}

func (s *SensorAlertService) handleMessage(topic string, payload []byte) {
	value, err := strconv.ParseFloat(string(payload), 64)
	if err != nil {
		log.Printf("alert: cannot parse payload for %s: %v", topic, err)
		return
	}
	for _, st := range s.byTopic[topic] {
		s.evaluate(st, value)
	}
}

func (s *SensorAlertService) evaluate(st *sensorAlertState, value float64) {
	inAlert, side := s.condition(st, value)

	if !st.haveSample {
		st.haveSample = true
		st.triggered = inAlert
		st.triggerSide = side
		if inAlert {
			s.fire(st, value)
		}
		return
	}

	if inAlert && !st.triggered {
		st.triggered = true
		st.triggerSide = side
		s.fire(st, value)
	} else if !inAlert && st.triggered {
		st.triggered = false
		st.triggerSide = alertSideNone
		s.dismiss(st)
	}
}

func (s *SensorAlertService) condition(st *sensorAlertState, value float64) (bool, sensorAlertSide) {
	cfg := st.cfg
	h := cfg.Hysteresis

	if st.triggered {
		// Currently in alarm — apply hysteresis on the side that fired so we
		// don't flicker right at the edge.
		switch st.triggerSide {
		case alertSideMin:
			if cfg.Min != nil && value < *cfg.Min+h {
				return true, alertSideMin
			}
		case alertSideMax:
			if cfg.Max != nil && value > *cfg.Max-h {
				return true, alertSideMax
			}
		}
		return false, alertSideNone
	}

	if cfg.Min != nil && value < *cfg.Min {
		return true, alertSideMin
	}
	if cfg.Max != nil && value > *cfg.Max {
		return true, alertSideMax
	}
	return false, alertSideNone
}

func (s *SensorAlertService) fire(st *sensorAlertState, value float64) {
	log.Printf("alert %s: fired (side=%s value=%v)", st.cfg.Name, st.triggerSide, value)

	if st.cfg.Sound != "" {
		if b, ok := s.sounds[st.cfg.Sound]; ok {
			go s.playSound(b)
		}
	}

	if game == nil {
		return
	}

	msg := st.cfg.Message
	if msg == "" {
		msg = st.cfg.Name
	}
	m := &ModalUi{
		stackLayout: []UiElement{
			&AlertUi{
				msg:  msg,
				icon: st.cfg.Icon,
			},
		},
	}
	m.Init()
	st.activeModal = m
	game.currentModal = m

	duration := st.cfg.DurationS
	if duration <= 0 {
		duration = 10
	}
	go func() {
		time.Sleep(time.Duration(duration) * time.Second)
		if game != nil && game.currentModal == m {
			game.currentModal = nil
		}
		if st.activeModal == m {
			st.activeModal = nil
		}
	}()
}

func (s *SensorAlertService) dismiss(st *sensorAlertState) {
	log.Printf("alert %s: cleared", st.cfg.Name)
	m := st.activeModal
	st.activeModal = nil
	if game != nil && m != nil && game.currentModal == m {
		game.currentModal = nil
	}
}

func (s *SensorAlertService) playSound(b []byte) {
	stream, err := mp3.DecodeWithSampleRate(44100, bytes.NewReader(b))
	if err != nil {
		log.Println("alert: decode:", err)
		return
	}
	player, err := s.audioContext.NewPlayer(stream)
	if err != nil {
		log.Println("alert: new player:", err)
		return
	}
	player.SetVolume(1)
	player.Play()
	go func() {
		for player.IsPlaying() {
			time.Sleep(200 * time.Millisecond)
		}
		if err := player.Close(); err != nil {
			log.Println("alert: player close:", err)
		}
	}()
}
