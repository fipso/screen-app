package main

import (
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttService struct {
	Client mqtt.Client
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
