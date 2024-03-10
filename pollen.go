package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type PollenUi struct {
	screen *ebiten.Image
}

type PollenData struct {
	Name       string `json:"name"`
	NextUpdate string `json:"next_update"`
	Legend     struct {
		ID4     string `json:"id4"`
		ID7Desc string `json:"id7_desc"`
		ID1Desc string `json:"id1_desc"`
		ID2Desc string `json:"id2_desc"`
		ID2     string `json:"id2"`
		ID5     string `json:"id5"`
		ID6Desc string `json:"id6_desc"`
		ID1     string `json:"id1"`
		ID4Desc string `json:"id4_desc"`
		ID7     string `json:"id7"`
		ID3Desc string `json:"id3_desc"`
		ID6     string `json:"id6"`
		ID3     string `json:"id3"`
		ID5Desc string `json:"id5_desc"`
	} `json:"legend"`
	Sender  string `json:"sender"`
	Content []struct {
		PartregionName string `json:"partregion_name"`
		RegionID       int    `json:"region_id"`
		RegionName     string `json:"region_name"`
		PartregionID   int    `json:"partregion_id"`
		Pollen         struct {
			Roggen struct {
				Today      string `json:"today"`
				Tomorrow   string `json:"tomorrow"`
				DayafterTo string `json:"dayafter_to"`
			} `json:"Roggen"`
			Graeser struct {
				Today      string `json:"today"`
				DayafterTo string `json:"dayafter_to"`
				Tomorrow   string `json:"tomorrow"`
			} `json:"Graeser"`
			Esche struct {
				Today      string `json:"today"`
				DayafterTo string `json:"dayafter_to"`
				Tomorrow   string `json:"tomorrow"`
			} `json:"Esche"`
			Erle struct {
				Today      string `json:"today"`
				Tomorrow   string `json:"tomorrow"`
				DayafterTo string `json:"dayafter_to"`
			} `json:"Erle"`
			Ambrosia struct {
				DayafterTo string `json:"dayafter_to"`
				Tomorrow   string `json:"tomorrow"`
				Today      string `json:"today"`
			} `json:"Ambrosia"`
			Hasel struct {
				DayafterTo string `json:"dayafter_to"`
				Tomorrow   string `json:"tomorrow"`
				Today      string `json:"today"`
			} `json:"Hasel"`
			Birke struct {
				DayafterTo string `json:"dayafter_to"`
				Tomorrow   string `json:"tomorrow"`
				Today      string `json:"today"`
			} `json:"Birke"`
			Beifuss struct {
				Today      string `json:"today"`
				Tomorrow   string `json:"tomorrow"`
				DayafterTo string `json:"dayafter_to"`
			} `json:"Beifuss"`
		} `json:"Pollen"`
	} `json:"content"`
	LastUpdate string `json:"last_update"`
}

var pollenStrength map[string]string

func pollPollen() {
	pollenStrength = make(map[string]string)

	for {
		fetchPollen()
		time.Sleep(10 * time.Minute)
	}
}

func fetchPollen() error {
	res, err := http.Get("https://opendata.dwd.de/climate_environment/health/alerts/s31fg.json")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var data PollenData
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	entry := data.Content[0]
	for _, region := range data.Content {
		if region.PartregionID == 43 {
			entry = region
			break
		}
	}

	pollenStrength["G"] = entry.Pollen.Graeser.Today
	pollenStrength["B"] = entry.Pollen.Birke.Today
	pollenStrength["H"] = entry.Pollen.Hasel.Today

	return nil
}

func (ui *PollenUi) Init() {
	ui.screen = ebiten.NewImage(WIDTH, fontHeight+linePadding)
}

func (ui *PollenUi) Draw() *ebiten.Image {
        ui.screen.Fill(bgColor)

	pollenS := "Pollen: "
	pollenKeys := []string{"G", "B", "H"}
	for _, key := range pollenKeys {
		v := pollenStrength[key]
		pollenS += fmt.Sprintf("%s%s ", key, v)
	}
	text.Draw(ui.screen, pollenS, defaultFont, 0, fontHeight, textColor)

	return ui.screen
}
