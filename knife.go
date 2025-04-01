package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

var attackRecords *KnifeAttackRes

type KnifeAttackUi struct {
	screen *ebiten.Image
}

type KnifeAttackRes struct {
	Items []struct {
		Date    string `json:"date"`
		GeoData struct {
			Components []struct {
				LongName  string   `json:"long_name"`
				ShortName string   `json:"short_name"`
				Types     []string `json:"types"`
			} `json:"components"`
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geoData"`
		ID       string `json:"id"`
		Link     string `json:"link"`
		Location string `json:"location"`
		Title    string `json:"title"`
		Wounded  bool   `json:"wounded"`
	} `json:"items"`
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

func pollKnifeAttacks() {
	for {
		fetchKnifeAttacks()
		time.Sleep(2 * time.Minute)
	}
}

func fetchKnifeAttacks() {
	// date format 2024-11-04 23:00:00.000Z
	startDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05.000Z")
	endDate := time.Now().Format("2006-01-02 15:04:05.000Z")
	filterRaw := fmt.Sprintf("date >= '%s' && date <= '%s' && geoData != null", startDate, endDate)
	// urlencode filter
	filter := url.QueryEscape(filterRaw)
	//log.Println("filter", filter)

	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		"https://messerinzidenz.de/api/collections/incidents/records?page=1&perPage=500&skipTotal=1&fields=id%2Ctitle%2CgeoData%2Cdate%2Clink%2Clocation%2Cwounded&filter="+filter,
		nil,
	)
	if err != nil {
		log.Println("Could build knife attacks req", err)
		return
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set(
		"user-agent",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
	)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not fetch knife attacks JSON", err)
		return
	}
	defer resp.Body.Close()
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Could not fetch knife attacks JSON", err)
		return
	}
	var data KnifeAttackRes
	err = json.Unmarshal(bodyData, &data)
	if err != nil {
		log.Println("Could not fetch knife attacks JSON", err)
		return
	}

	attackRecords = &data
}

func (ui *KnifeAttackUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *KnifeAttackUi) Bounds() (width, height int) {
	return config.Width, 800
}

func (ui *KnifeAttackUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	text.Draw(
		ui.screen,
		fmt.Sprintf(
			" messerinzidenz  %d",
			len(attackRecords.Items),
		),
		defaultFont,
		0,
		100,
		textColor,
	)

	height := 0
	for _, attack := range attackRecords.Items {
		c := textColor
		if attack.Wounded {
			c = color.RGBA{255, 0, 0, 255}
		}

		var t string
		if len(attack.Title)+len(attack.Location) < 40 {
			t = fmt.Sprintf(
				"%s - %s",
				attack.Location,
				attack.Title,
			)
			text.Draw(
				ui.screen,
				t,
				smallFont,
				0,
				190+height,
				c,
			)
			height += 48 + linePadding
		} else {
			t = fmt.Sprintf(
				"%s:",
				attack.Location,
			)
			text.Draw(
				ui.screen,
				t,
				smallFont,
				0,
				190+height,
				c,
			)
			height += 48
			text.Draw(
				ui.screen,
				attack.Title,
				smallFont,
				20,
				190+height,
				c,
			)
			height += 48 + linePadding
		}
	}

	return ui.screen
}
