package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type WeatherUi struct {
	screen *ebiten.Image
}

func pollWeather() {
	for {
		fetchPollen()
		time.Sleep(10 * time.Minute)
	}
}

func fetchWeather() error {
	// res, err := http.Get("https://opendata.dwd.de/climate_environment/health/alerts/s31fg.json")
	// if err != nil {
	// 	return err
	// }
	// defer res.Body.Close()
	// b, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	return err
	// }
	// var data PollenData
	// err = json.Unmarshal(b, &data)
	// if err != nil {
	// 	return err
	// }

	// entry := data.Content[0]
	// for _, region := range data.Content {
	// 	if region.PartregionID == 43 {
	// 		entry = region
	// 		break
	// 	}
	// }

	// pollenStrength["G"] = entry.Pollen.Graeser.Today
	// pollenStrength["B"] = entry.Pollen.Birke.Today
	// pollenStrength["H"] = entry.Pollen.Hasel.Today

	return nil
}

func (ui *WeatherUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *WeatherUi) Bounds() (width, height int) {
	return WIDTH, fontHeight + linePadding
}

func (ui *WeatherUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	weatherS := "Weather: ???"
	text.Draw(ui.screen, weatherS, defaultFont, 0, fontHeight, textColor)

	return ui.screen
}
