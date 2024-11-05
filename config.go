package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
)

type Config struct {
	Fullscreen bool
	Width      int
	Height     int
	Grow_mqtt  struct {
		Enabled    bool
		Server     string
		Box_temp   string
		Box_humid  string
		Room_temp  string
		Room_humid string
	}
}

func loadConfig() {
	configPath := "config.json"
	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	// Create config if not exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config = Config{}
		// Save empty config to file
		configB, err := json.Marshal(config)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(configPath, configB, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Load config from file
	configB, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(configB, &config)
	if err != nil {
		log.Fatal(err)
	}
        spew.Dump(config)
}
