package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Fullscreen bool
	Width      int
	Height     int
	Mqtt       struct {
		Username string
		Password string
		Enabled  bool
		Server   string
	}
	Grow struct {
		Sensors []struct {
			Name  string
			Temp  string
			Humid string
		}
	}
	Default_Font_Size int
	Layout            []LayoutElement
	Energy            struct {
		MaxHistoryHours int
		Profiles        map[string]string
		Devices         []RefossEnergyDeviceConfig
	}
}

type RefossEnergyDeviceConfig struct {
	Name    string
	Address string
	UUID    string
	Profile string
}

type LayoutElement struct {
	Type           LayoutElementType
	SwitchInterval int
	Children       []LayoutElement
}

type LayoutElementType string

const (
	LayoutElementSwitch  = LayoutElementType("switch")
	LayoutElementGrow    = LayoutElementType("grow")
	LayoutElementBus     = LayoutElementType("bus")
	LayoutElementWeather = LayoutElementType("weather")
	LayoutElementKnife   = LayoutElementType("knife")
	LayoutElementClock   = LayoutElementType("clock")
	LayoutElementCrypto  = LayoutElementType("crypto")
	LayoutElementEnergy  = LayoutElementType("energy")
)

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
		err = os.WriteFile(configPath, configB, 0o644)
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

	// Apply default
	if config.Width == 0 {
		config.Width = 800
	}
	if config.Height == 0 {
		config.Height = 600
	}
	if config.Default_Font_Size == 0 {
		config.Default_Font_Size = 72
	}
	if config.Energy.MaxHistoryHours == 0 {
		config.Energy.MaxHistoryHours = 6
	}
}
