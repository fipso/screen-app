package main

import (
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

type UiElement interface {
	Init()
	Draw() *ebiten.Image
}

// Ebiten units
const WIDTH = 360 / 2
const HEIGHT = 640 / 2

var textColor = color.RGBA{255, 255, 255, 255}
var bgColor = color.RGBA{0, 0, 0, 255}
var defaultFont font.Face = basicfont.Face7x13
var fontHeight = 10
var linePadding = 2

type Game struct {
	stackLayout []UiElement
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(bgColor)

	// Draw UI elements
	pos := ebiten.GeoM{}
	pos.Translate(10, 0)
	for _, ui := range g.stackLayout {
		img := ui.Draw()
		screen.DrawImage(img, &ebiten.DrawImageOptions{
			GeoM: pos,
		})

		// Move to the next position
		pos.Translate(0, float64(img.Bounds().Dy()+(linePadding)))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WIDTH, HEIGHT
}

func runGameUI() {
	game := &Game{}

	// Load UI elements
	game.stackLayout = append(game.stackLayout, &ClockUi{})
	game.stackLayout = append(game.stackLayout, &CryptoUi{})
	game.stackLayout = append(game.stackLayout, &BusUi{})
	game.stackLayout = append(game.stackLayout, &PollenUi{})

	for _, ui := range game.stackLayout {
		ui.Init()
	}

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
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
