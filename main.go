package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"runtime/pprof"
)

var (
	game        *Game
	config      Config
	mqttService MqttService
)

func main() {
	// cli flags
	cliLayout := flag.String("layout", "", "override config layout from cli")
	// profiling flags
	cpuProfile := flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile := flag.String("memprofile", "", "write memory profile to file")
	flag.Parse()

	// Start CPU profiling if flag is provided
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Write memory profile if flag is provided
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

	loadConfig()

	mqttService = MqttService{}
	go mqttService.Run()

	doorService := DoorService{}
	go doorService.Run()

	go pollBinance()
	go pollBusTimes()
	go pollPollen()
	go pollWeather()
	go pollKnifeAttacks()

	// Override layout from cli if provided
	if *cliLayout != "" {
		err := json.Unmarshal([]byte(*cliLayout), &config.Layout)
		if err != nil {
			log.Fatal("could not parse layout from cli: ", err)
		}
	}

	runGameUI()
}
