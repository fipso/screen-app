package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	prices := sortedPrices()
	for i, currency := range prices {
		c := color.RGBA{255, 255, 255, 255}
		delta := oneHourDelta(currency)
		if delta > 0 {
			c = color.RGBA{0, 255, 0, 255}
		} else if delta < 0 {
			c = color.RGBA{255, 0, 0, 255}
		}
		text.Draw(screen, fmt.Sprintf("%-9s: %.4f", currency.name, currency.price), basicfont.Face7x13, 20, 20+16*i, c)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 360, 640
}

func runGameUI() {
	ebiten.SetWindowSize(1080, 1920)
	//ebiten.SetWindowSize(1080 / 4, 1920 / 4)
	ebiten.SetWindowTitle("Screep App Game UI")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
