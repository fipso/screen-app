package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var loc *time.Location

func main() {
	var err error
	loc, err = time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Fatal(err)
	}

	for {
		txt := "784er:\n\n"
		txt += GetBusTime("-> Night City", "de%3A05158%3A19001", "de%3A05158%3A13980")
		txt += "\n"
		txt += GetBusTime("-> Die Rosi", "de%3A05158%3A19001", "de%3A05158%3A18969")
		fmt.Print("\033[H\033[2J")
		fmt.Print(txt)
		time.Sleep(time.Second * 30)
	}

}

func GetBusTime(label, origin, destination string) string {
	resp, err := http.Get(fmt.Sprintf(
		"https://www.vrr.de/vrr-efa/XML_TRIP_REQUEST2?allInterchangesAsLegs=1&calcOneDirection=1&changeSpeed=normal&convertAddressesITKernel2LocationServer=1&convertCoord2LocationServer=1&convertCrossingsITKernel2LocationServer=1&convertPOIsITKernel2LocationServer=1&convertStopsPTKernel2LocationServer=1&coordOutputDistance=1&coordOutputFormat=WGS84%%5Bdd.ddddd%%5D&genC=1&genMaps=0&imparedOptionsActive=1&inclMOT_0=true&inclMOT_1=true&inclMOT_10=true&inclMOT_11=true&inclMOT_12=true&inclMOT_13=true&inclMOT_17=true&inclMOT_18=true&inclMOT_19=true&inclMOT_2=true&inclMOT_3=true&inclMOT_4=true&inclMOT_5=true&inclMOT_6=true&inclMOT_7=true&inclMOT_8=true&inclMOT_9=true&includedMeans=checkbox&itOptionsActive=1&itdTripDateTimeDepArr=dep&language=de&lineRestriction=400&locationServerActive=1&maxChanges=9&name_destination=%s&name_origin=%s&outputFormat=rapidJSON&ptOptionsActive=1&routeType=LEASTTIME&serverInfo=1&sl3plusTripMacro=1&trITMOTvalue100=10&type_destination=any&type_notVia=any&type_origin=any&type_via=any&useElevationData=1&useProxFootSearch=true&useRealtime=1&useUT=1&version=10.5.17.3&vrrTripMacro=1",
		destination,
		origin,
	))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var busData BusData
	err = json.Unmarshal(b, &busData)
	if err != nil {
		log.Fatal(err)
	}

	s := label
	n := 0
	for _, journey := range busData.Journeys {
		// Skip wrong bus
		conn := journey.Legs[0]
		if conn.Transportation.Number != "784" {
			continue
		}

		// Set ansi color
		if conn.Origin.DepartureTimeEstimated.Sub(conn.Origin.DepartureTimePlanned).Minutes() > 3 {
			s += "\033[31m"
		} else {
			s += "\033[32m"
		}
		s += "\n    "
		s += conn.Origin.DepartureTimeEstimated.In(loc).Format("15:04")
		// Reset ansi color
		s += "\033[0m"
		if n == 2 {
			break
		}
		n++
	}

        s += "\n"

	return s
}
