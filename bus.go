package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type BusUi struct {
	screen *ebiten.Image
}

type BusData struct {
	ServerInfo struct {
		ControllerVersion string  `json:"controllerVersion"`
		ServerID          string  `json:"serverID"`
		VirtDir           string  `json:"virtDir"`
		ServerTime        string  `json:"serverTime"`
		CalcTime          float64 `json:"calcTime"`
	} `json:"serverInfo"`
	Version        string `json:"version"`
	SystemMessages []struct {
		Type    string `json:"type"`
		Module  string `json:"module"`
		Code    int    `json:"code"`
		Text    string `json:"text"`
		SubType string `json:"subType"`
	} `json:"systemMessages"`
	Journeys []struct {
		Rating       int  `json:"rating"`
		IsAdditional bool `json:"isAdditional"`
		Interchanges int  `json:"interchanges"`
		Legs         []struct {
			Duration             int      `json:"duration"`
			IsRealtimeControlled bool     `json:"isRealtimeControlled"`
			RealtimeStatus       []string `json:"realtimeStatus"`
			Origin               struct {
				IsGlobalID       bool      `json:"isGlobalId"`
				ID               string    `json:"id"`
				Name             string    `json:"name"`
				DisassembledName string    `json:"disassembledName"`
				Type             string    `json:"type"`
				Coord            []float64 `json:"coord"`
				Niveau           int       `json:"niveau"`
				Parent           struct {
					IsGlobalID       bool   `json:"isGlobalId"`
					ID               string `json:"id"`
					Name             string `json:"name"`
					DisassembledName string `json:"disassembledName"`
					Type             string `json:"type"`
					Parent           struct {
						ID   string `json:"id"`
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"parent"`
					Properties struct {
						StopID string `json:"stopId"`
					} `json:"properties"`
					Coord  []float64 `json:"coord"`
					Niveau int       `json:"niveau"`
				} `json:"parent"`
				ProductClasses         []int     `json:"productClasses"`
				DepartureTimePlanned   time.Time `json:"departureTimePlanned"`
				DepartureTimeEstimated time.Time `json:"departureTimeEstimated"`
				Properties             struct {
					AREANIVEAUDIVA       string `json:"AREA_NIVEAU_DIVA"`
					StoppingPointPlanned string `json:"stoppingPointPlanned"`
					AreaGid              string `json:"areaGid"`
					Area                 string `json:"area"`
					Platform             string `json:"platform"`
					PlatformName         string `json:"platformName"`
				} `json:"properties"`
			} `json:"origin"`
			Destination struct {
				IsGlobalID       bool      `json:"isGlobalId"`
				ID               string    `json:"id"`
				Name             string    `json:"name"`
				DisassembledName string    `json:"disassembledName"`
				Type             string    `json:"type"`
				Coord            []float64 `json:"coord"`
				Niveau           int       `json:"niveau"`
				Parent           struct {
					IsGlobalID bool   `json:"isGlobalId"`
					ID         string `json:"id"`
					Name       string `json:"name"`
					Type       string `json:"type"`
					Parent     struct {
						ID   string `json:"id"`
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"parent"`
					Properties struct {
						StopID string `json:"stopId"`
					} `json:"properties"`
					Coord  []float64 `json:"coord"`
					Niveau int       `json:"niveau"`
				} `json:"parent"`
				ProductClasses       []int     `json:"productClasses"`
				ArrivalTimePlanned   time.Time `json:"arrivalTimePlanned"`
				ArrivalTimeEstimated time.Time `json:"arrivalTimeEstimated"`
				Properties           struct {
					AREANIVEAUDIVA       string `json:"AREA_NIVEAU_DIVA"`
					StoppingPointPlanned string `json:"stoppingPointPlanned"`
					AreaGid              string `json:"areaGid"`
					Area                 string `json:"area"`
					Platform             string `json:"platform"`
					PlatformName         string `json:"platformName"`
				} `json:"properties"`
			} `json:"destination"`
			Transportation struct {
				ID               string `json:"id"`
				Name             string `json:"name"`
				DisassembledName string `json:"disassembledName"`
				Number           string `json:"number"`
				Description      string `json:"description"`
				Product          struct {
					ID     int    `json:"id"`
					Class  int    `json:"class"`
					Name   string `json:"name"`
					IconID int    `json:"iconId"`
				} `json:"product"`
				Operator struct {
					Code string `json:"code"`
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"operator"`
				Destination struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
				} `json:"destination"`
				Properties struct {
					IsSTT           bool   `json:"isSTT"`
					IsROP           bool   `json:"isROP"`
					TripCode        int    `json:"tripCode"`
					TimetablePeriod string `json:"timetablePeriod"`
					LineDisplay     string `json:"lineDisplay"`
					GlobalID        string `json:"globalId"`
					AVMSTripID      string `json:"AVMSTripID"`
				} `json:"properties"`
			} `json:"transportation"`
			StopSequence []struct {
				IsGlobalID       bool      `json:"isGlobalId"`
				ID               string    `json:"id"`
				Name             string    `json:"name"`
				DisassembledName string    `json:"disassembledName"`
				Type             string    `json:"type"`
				Coord            []float64 `json:"coord"`
				Niveau           int       `json:"niveau"`
				Parent           struct {
					IsGlobalID       bool   `json:"isGlobalId"`
					ID               string `json:"id"`
					Name             string `json:"name"`
					DisassembledName string `json:"disassembledName"`
					Type             string `json:"type"`
					Parent           struct {
						ID   string `json:"id"`
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"parent"`
					Properties struct {
						StopID string `json:"stopId"`
					} `json:"properties"`
					Coord  []float64 `json:"coord"`
					Niveau int       `json:"niveau"`
				} `json:"parent"`
				ProductClasses []int `json:"productClasses"`
				Properties     struct {
					AREANIVEAUDIVA       string `json:"AREA_NIVEAU_DIVA"`
					StoppingPointPlanned string `json:"stoppingPointPlanned"`
					AreaGid              string `json:"areaGid"`
					Area                 string `json:"area"`
					Platform             string `json:"platform"`
					PlatformName         string `json:"platformName"`
					Zone                 string `json:"zone"`
				} `json:"properties"`
				DepartureTimePlanned   time.Time `json:"departureTimePlanned,omitempty"`
				DepartureTimeEstimated time.Time `json:"departureTimeEstimated,omitempty"`
				ArrivalTimePlanned     time.Time `json:"arrivalTimePlanned,omitempty"`
				ArrivalTimeEstimated   time.Time `json:"arrivalTimeEstimated,omitempty"`
			} `json:"stopSequence"`
			Infos []struct {
				Priority   string `json:"priority"`
				ID         string `json:"id"`
				Version    int    `json:"version"`
				Type       string `json:"type"`
				Properties struct {
					InfoType         string `json:"infoType"`
					IncidentDateTime string `json:"incidentDateTime"`
				} `json:"properties"`
				InfoLinks []struct {
					URLText  string `json:"urlText"`
					URL      string `json:"url"`
					Content  string `json:"content"`
					Subtitle string `json:"subtitle"`
					Title    string `json:"title"`
					WapText  string `json:"wapText"`
					SmsText  string `json:"smsText"`
				} `json:"infoLinks"`
			} `json:"infos"`
			Coords [][]interface{} `json:"coords"`
			Fare   struct {
				Zones []struct {
					Net         string        `json:"net"`
					ToLeg       int           `json:"toLeg"`
					FromLeg     int           `json:"fromLeg"`
					NeutralZone string        `json:"neutralZone"`
					Zones       []interface{} `json:"zones"`
					ZonesUnited [][]string    `json:"zonesUnited"`
				} `json:"zones"`
			} `json:"fare"`
			Properties struct {
				VehicleAccess        []string `json:"vehicleAccess"`
				PlanWheelChairAccess string   `json:"PlanWheelChairAccess"`
			} `json:"properties"`
		} `json:"legs"`
		Fare struct {
			Tickets []struct {
				ID                      string  `json:"id"`
				Name                    string  `json:"name"`
				Comment                 string  `json:"comment"`
				URL                     string  `json:"URL"`
				Currency                string  `json:"currency"`
				PriceLevel              string  `json:"priceLevel"`
				PriceBrutto             float64 `json:"priceBrutto"`
				PriceNetto              float64 `json:"priceNetto"`
				TaxPercent              float64 `json:"taxPercent"`
				FromLeg                 int     `json:"fromLeg"`
				ToLeg                   int     `json:"toLeg"`
				Net                     string  `json:"net"`
				Person                  string  `json:"person"`
				TravellerClass          string  `json:"travellerClass"`
				TimeValidity            string  `json:"timeValidity"`
				ValidMinutes            int     `json:"validMinutes"`
				IsShortHaul             string  `json:"isShortHaul"`
				ReturnsAllowed          string  `json:"returnsAllowed"`
				ValidForOneJourneyOnly  string  `json:"validForOneJourneyOnly"`
				ValidForOneOperatorOnly string  `json:"validForOneOperatorOnly"`
				NumberOfChanges         int     `json:"numberOfChanges"`
				NameValidityArea        string  `json:"nameValidityArea"`
				RelationKeys            []struct {
					ID   string `json:"id"`
					Code string `json:"code"`
					Name string `json:"name"`
				} `json:"relationKeys,omitempty"`
				ValidFrom  time.Time `json:"validFrom,omitempty"`
				ValidTo    time.Time `json:"validTo,omitempty"`
				Properties struct {
					RiderCategoryName    string        `json:"riderCategoryName"`
					DisplayGroup         string        `json:"displayGroup"`
					TicketType           string        `json:"ticketType"`
					ProductID            string        `json:"productID"`
					ValidityStartDate    string        `json:"validity_start_date"`
					ValidityStartTime    string        `json:"validity_start_time"`
					ValidityEndDate      string        `json:"validity_end_date"`
					ValidityEndTime      string        `json:"validity_end_time"`
					DistExact            int           `json:"distExact"`
					DistRounded          int           `json:"distRounded"`
					PricePerKM           float64       `json:"pricePerKM"`
					PriceBasic           float64       `json:"priceBasic"`
					TariffProductDefault []interface{} `json:"tariffProductDefault"`
					TariffProductOption  []interface{} `json:"tariffProductOption"`
				} `json:"properties"`
				TargetGroups []string `json:"targetGroups,omitempty"`
			} `json:"tickets"`
			Zones []struct {
				Net         string     `json:"net"`
				ToLeg       int        `json:"toLeg"`
				FromLeg     int        `json:"fromLeg"`
				NeutralZone string     `json:"neutralZone"`
				Zones       [][]string `json:"zones"`
				ZonesUnited [][]string `json:"zonesUnited"`
			} `json:"zones"`
		} `json:"fare"`
		DaysOfService struct {
			Rvb string `json:"rvb"`
		} `json:"daysOfService"`
	} `json:"journeys"`
}

type busTime struct {
	time  time.Time
	delay time.Duration
}

var loc *time.Location
var busTimes = map[string][]busTime{}

func pollBusTimes() {
	var err error
	loc, err = time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Fatal(err)
	}

	for {
		busTimes["W. Tal"] = GetBusTime("de%3A05158%3A19001", "de%3A05158%3A13980")
		busTimes["D. Dorf"] = GetBusTime("de%3A05158%3A19001", "de%3A05158%3A18969")
		time.Sleep(time.Second * 30)
	}

}

func GetBusTime(origin, destination string) []busTime {
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

	var res []busTime

JOURNEY_LOOP:
	for _, journey := range busData.Journeys {
		// Skip wrong bus & multi bus connections
		for _, leg := range journey.Legs {
			if leg.Transportation.Number != "784" {
				continue JOURNEY_LOOP
			}
		}

		leg := journey.Legs[0]

		// Set ansi color
		delay := leg.Origin.DepartureTimeEstimated.Sub(leg.Origin.DepartureTimePlanned)
		localTime := leg.Origin.DepartureTimeEstimated.In(loc)

		res = append(res, busTime{time: localTime, delay: delay})
	}

	return res
}

func (ui *BusUi) Init() {
	ui.screen = ebiten.NewImage(WIDTH, (fontHeight+linePadding)*7)
}

func (ui *BusUi) Draw() *ebiten.Image {
        ui.screen.Fill(bgColor)

	busKeys := []string{"W. Tal", "D. Dorf"}
	for i, key := range busKeys {
		text.Draw(ui.screen, key, defaultFont, 64*i, fontHeight, textColor)
		times := busTimes[key]
		for j, entry := range times {
			c := textColor
			if entry.delay.Minutes() > 3 {
				c = color.RGBA{255, 0, 0, 255}
			}

			text.Draw(ui.screen, entry.time.Format("15:04"), defaultFont, 64*i, (fontHeight+linePadding)*(j+2), c)
		}
	}

	return ui.screen
}
