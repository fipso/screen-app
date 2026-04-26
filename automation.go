package main

import (
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type AutomationService struct {
	states  map[string]*automationState
	byTopic map[string][]*automationState
}

type automationState struct {
	cfg        AutomationConfig
	device     *RefossEnergyDeviceConfig
	triggered  bool
	haveSample bool
}

func (s *AutomationService) Run() {
	mqttService.WaitReady()

	s.states = map[string]*automationState{}
	s.byTopic = map[string][]*automationState{}

	for _, cfg := range config.Automations {
		var device *RefossEnergyDeviceConfig
		for i := range config.Energy.Devices {
			if config.Energy.Devices[i].UUID == cfg.DeviceUUID {
				device = &config.Energy.Devices[i]
				break
			}
		}
		if device == nil {
			log.Printf("automation %s: device with UUID %s not found", cfg.Name, cfg.DeviceUUID)
			continue
		}

		st := &automationState{cfg: cfg, device: device}
		s.states[cfg.Name] = st
		s.byTopic[cfg.Topic] = append(s.byTopic[cfg.Topic], st)
	}

	for topic, sts := range s.byTopic {
		t := topic
		names := make([]string, 0, len(sts))
		for _, st := range sts {
			names = append(names, st.cfg.Name)
		}
		log.Printf("automation: subscribing to %s for %v", t, names)
		mqttService.On(t, func(client mqtt.Client, msg mqtt.Message) {
			s.handleMessage(t, msg.Payload())
		})
	}
}

func (s *AutomationService) handleMessage(topic string, payload []byte) {
	value, err := strconv.ParseFloat(string(payload), 64)
	if err != nil {
		log.Printf("automation: cannot parse payload for %s: %v", topic, err)
		return
	}
	log.Printf("automation: %s = %v", topic, value)
	for _, st := range s.byTopic[topic] {
		s.evaluate(st, value)
	}
}

func (s *AutomationService) evaluate(st *automationState, value float64) {
	cond := s.condition(st, value)

	if !st.haveSample {
		st.haveSample = true
		st.triggered = cond
		s.fire(st, cond)
		return
	}

	if cond != st.triggered {
		st.triggered = cond
		s.fire(st, cond)
	}
}

func (s *AutomationService) condition(st *automationState, value float64) bool {
	threshold := st.cfg.Threshold
	switch st.cfg.Operator {
	case AutomationOpAbove:
		if st.triggered {
			return value > threshold-st.cfg.Hysteresis
		}
		return value > threshold
	case AutomationOpBelow:
		if st.triggered {
			return value < threshold+st.cfg.Hysteresis
		}
		return value < threshold
	}
	return false
}

func (s *AutomationService) fire(st *automationState, cond bool) {
	want := st.cfg.OnTrigger
	if !cond {
		want = !st.cfg.OnTrigger
	}
	log.Printf("automation %s: condition=%v -> setting %s (%s) to %v", st.cfg.Name, cond, st.device.Name, st.device.UUID, want)
	go func() {
		if err := st.device.SetPlugState(want); err != nil {
			log.Printf("automation %s: SetPlugState(%v) failed: %v", st.cfg.Name, want, err)
		}
	}()
}
