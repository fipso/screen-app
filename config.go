package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	fullscreen bool
	grow_mqtt  struct {
		enabled    bool
		server     string
		box_temp   string
		box_humid  string
		room_temp  string
		room_humid string
	}
}

func loadConfig() {
	// Create config if not exists
	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		config = &Config{}
		// Save empty config to file
		configB, err := json.Marshal(config)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile("config.json", configB, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Load config from file
	configB, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(configB, &config)
	if err != nil {
		log.Fatal(err)
	}
}
