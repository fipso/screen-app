package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, pricesText)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 576, 1024
}

func RunGameUI() {
	ebiten.SetWindowSize(1080, 1920)
	ebiten.SetWindowTitle("Screep App Game UI")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
