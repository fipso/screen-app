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
	"golang.org/x/net/html"
)

var attackRecords *KnifeAttackRes
var focusedAttackIndex int
var attackDetails map[string]string

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

	attackRecords = &data
}

func fetchAttackDetails(url string) (string, error) {
	// Please crawl the url and extract the main text section
	req, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	// Find main text element in HTML
	doc, err := html.Parse(req.Body)
	if err != nil {
		return "", err
	}

	var mainText string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "main" {
			mainText = extractText(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return mainText, nil
}

func extractText(n *html.Node) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		} else if c.Type == html.ElementNode {
			text += extractText(c)
		}
	}
	return strings.TrimSpace(text)
}

func (ui *KnifeAttackUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *KnifeAttackUi) Bounds() (width, height int) {
	return WIDTH, 800
}

func (ui *KnifeAttackUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	text.Draw(
		ui.screen,
		fmt.Sprintf(
			"messerinzidenz %d",
			len(attackRecords.Items),
		),
		defaultFont,
		0,
		100,
		textColor,
	)

	for i, attack := range attackRecords.Items {
		c := textColor
		if attack.Wounded {
			c = color.RGBA{255, 0, 0, 255}
		}

		text.Draw(
			ui.screen,
			fmt.Sprintf(
				"%s - %s",
				attack.Location,
				attack.Title,
			),
			smallFont,
			0,
			120+48*(i+1),
			c,
		)
	}

	return ui.screen
}
