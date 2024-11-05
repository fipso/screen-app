package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type KnifeAttackUi struct {
	screen *ebiten.Image

	attackRecords *KnifeAttackRes
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

func (ui *KnifeAttackUi) fetchKnifeAttacks() {
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		"https://messerinzidenz.de/api/collections/incidents/records?page=1&perPage=500&skipTotal=1&fields=id%2Ctitle%2CgeoData%2Cdate%2Clink%2Clocation%2Cwounded&filter=date%20%3E%3D%20%272024-09-08%2022%3A00%3A00.000Z%27%20%26%26%20date%20%3C%3D%20%272024-09-09%2021%3A59%3A59.999Z%27%20%26%26%20geoData%20!%3D%20null",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set(
		"user-agent",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
	)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var data KnifeAttackRes
	err = json.Unmarshal(bodyData, &data)
	if err != nil {
		log.Fatal(err)
	}

	ui.attackRecords = &data
}

func (ui *KnifeAttackUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *KnifeAttackUi) Bounds() (width, height int) {
	return WIDTH, 1420
}

func (ui *KnifeAttackUi) Draw() *ebiten.Image {
	text.Draw(
		ui.screen,
		fmt.Sprintf(
			"Messerinzidenz %d",
			len(ui.attackRecords.Items),
		),
		defaultFont,
		0,
		0,
		textColor,
	)

	return ui.screen
}
