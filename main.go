package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
)

var config Config

func main() {
	// CPU profiling flag
	cpuProfile := flag.String("cpuprofile", "", "write cpu profile to file")
	// Memory profiling flag
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

	go pollBinance()
	go pollBusTimes()
	go pollPollen()
	go pollWeather()
	go pollKnifeAttacks()
	go pollEnergyDevices()

	loadConfig()
	runGameUI()

}
