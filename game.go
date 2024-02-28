package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

var textColor = color.RGBA{255, 255, 255, 255}
var bgColor = color.RGBA{0, 0, 0, 255}

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(bgColor)

	prices := sortedPrices()
	for i, currency := range prices {
		c := textColor
		delta := oneHourDelta(currency)
		if delta > 0 {
			c = color.RGBA{20, 200, 20, 255}
		} else if delta < 0 {
			c = color.RGBA{255, 0, 0, 255}
		}
		text.Draw(screen, fmt.Sprintf("%-9s: %.4f", currency.name, currency.price), basicfont.Face7x13, 20, 20+16*i, c)
	}

	offsetTop := 20 + 16*len(prices)
	busKeys := []string{"W. Tal", "D. Dorf"}
	for i, key := range busKeys {
		text.Draw(screen, key, basicfont.Face7x13, 20+64*i, offsetTop+16, textColor)
		times := busTimes[key]
		for j, entry := range times {
			c := textColor
			if entry.delay.Minutes() > 3 {
				c = color.RGBA{255, 0, 0, 255}
			}

			text.Draw(screen, entry.time.Format("15:04"), basicfont.Face7x13, 20+64*i, offsetTop+32+16*j, c)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 360 / SCALE, 640 / SCALE
}

func runGameUI() {
	// Dark/Light mode
	go func() {
		if time.Now().Hour() > 18 || time.Now().Hour() < 6 {
			textColor = color.RGBA{255, 255, 255, 255}
			bgColor = color.RGBA{0, 0, 0, 255}
		} else {
			textColor = color.RGBA{0, 0, 0, 255}
			bgColor = color.RGBA{245, 245, 245, 255}
		}

		time.Sleep(time.Second)
	}()

	ebiten.SetWindowSize(1080, 1920)
	ebiten.SetWindowTitle("Screep App Game UI")
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
