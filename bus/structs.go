package main

import "time"

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
