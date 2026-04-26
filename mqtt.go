package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttService struct {
	Client   mqtt.Client
	mu       sync.Mutex
	handlers map[string][]mqtt.MessageHandler
}

func (s *MqttService) Run() {
	// Connect to mqtt
	opts := mqtt.NewClientOptions()
	spew.Dump(config)
	opts.AddBroker(fmt.Sprintf("tcp://%s", config.Mqtt.Server))
	opts.SetClientID(fmt.Sprintf("screen-app-%d", time.Now().Unix()))
	opts.SetDefaultPublishHandler(s.defaultMessagePubHandler)
	opts.SetUsername(config.Mqtt.Username)
	opts.SetPassword(config.Mqtt.Password)
	s.Client = mqtt.NewClient(opts)
	if token := s.Client.Connect(); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		return
	}
	log.Println("Connected to MQTT")
}

func (s *MqttService) WaitReady() {
	for s.Client == nil || !s.Client.IsConnected() {
		time.Sleep(time.Millisecond * 50)
	}
}

func (s *MqttService) defaultMessagePubHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received unhandled message on topic: %s\nMessage: %s\n", msg.Topic(), msg.Payload())
}

// On registers a handler for a topic. Multiple handlers can be registered for the
// same topic — paho's Subscribe only keeps one handler per topic, so this routes
// through a single dispatcher.
func (s *MqttService) On(topic string, handler mqtt.MessageHandler) {
	s.mu.Lock()
	if s.handlers == nil {
		s.handlers = map[string][]mqtt.MessageHandler{}
	}
	_, exists := s.handlers[topic]
	s.handlers[topic] = append(s.handlers[topic], handler)
	s.mu.Unlock()

	if exists {
		return
	}
	s.Client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		s.mu.Lock()
		hs := s.handlers[msg.Topic()]
		s.mu.Unlock()
		for _, h := range hs {
			h(client, msg)
		}
	})
}
