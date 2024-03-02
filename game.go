package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

var textColor = color.RGBA{255, 255, 255, 255}
var bgColor = color.RGBA{0, 0, 0, 255}
var defaultFont font.Face = basicfont.Face7x13

var moscowLoc *time.Location
var washingtonLoc *time.Location

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(bgColor)

	offsetTop := 0

	// Time
	text.Draw(screen, time.Now().Format("15:04"), defaultFont, 20, offsetTop+16, textColor)
	text.Draw(screen, time.Now().In(moscowLoc).Format("15:04"), defaultFont, 70, offsetTop+16, textColor)
	text.Draw(screen, time.Now().In(washingtonLoc).Format("15:04"), defaultFont, 120, offsetTop+16, textColor)

	offsetTop += 18

	// Binance Ticker
	prices := sortedPrices()
	for i, currency := range prices {
		c := textColor
		delta := oneHourDelta(currency)
		if delta > 0 {
			c = color.RGBA{20, 200, 20, 255}
		} else if delta < 0 {
			c = color.RGBA{255, 0, 0, 255}
		}
		text.Draw(screen, fmt.Sprintf("%-9s: %.2f", currency.name, currency.price), defaultFont, 20, offsetTop+20+16*i, c)
	}

	// Bus Times
	offsetTop += 10 + 16*len(prices)
	busKeys := []string{"W. Tal", "D. Dorf"}
	for i, key := range busKeys {
		text.Draw(screen, key, defaultFont, 20+64*i, offsetTop+16, textColor)
		times := busTimes[key]
		for j, entry := range times {
			c := textColor
			if entry.delay.Minutes() > 3 {
				c = color.RGBA{255, 0, 0, 255}
			}

			text.Draw(screen, entry.time.Format("15:04"), basicfont.Face7x13, 20+64*i, offsetTop+32+16*j, c)
		}
	}

	// Pollen
	offsetTop += 155
	pollenS := "Pollen: "
	pollenKeys := []string{"G", "B", "H"}
	for _, key := range pollenKeys {
		v := pollenStrength[key]
		pollenS += fmt.Sprintf("%s%s ", key, v)
	}
	text.Draw(screen, pollenS, defaultFont, 20, offsetTop, textColor)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 360 / SCALE, 640 / SCALE
}

func runGameUI() {
	// Load locations
	moscowLoc, _ = time.LoadLocation("Europe/Moscow")
	washingtonLoc, _ = time.LoadLocation("America/New_York")

	// Dark/Light mode
	go func() {
		for {
			if time.Now().Hour() > 18 || time.Now().Hour() < 6 {
				textColor = color.RGBA{255, 255, 255, 255}
				bgColor = color.RGBA{0, 0, 0, 255}
			} else {
				textColor = color.RGBA{0, 0, 0, 255}
				bgColor = color.RGBA{245, 245, 245, 255}
			}

			time.Sleep(time.Second)
		}
	}()

	//Load font
	// fontData, err := os.ReadFile("assets/fonts/OpenSans-Regular.ttf")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// tt, err := opentype.Parse(fontData)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defaultFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
	// 	Size:    12,
	// 	DPI:     96,
	// 	Hinting: font.HintingFull,
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	//ebiten.SetWindowSize(1080, 1920)
	ebiten.SetWindowTitle("Screep App Game UI")
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
