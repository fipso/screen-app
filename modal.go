package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type ModalUi struct {
	screen      *ebiten.Image
	stackLayout []UiElement

	contentScreen *ebiten.Image
}

func (ui *ModalUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)

	ui.contentScreen = ebiten.NewImage(width, 400)

	for _, elem := range ui.stackLayout {
		elem.Init()
	}
}

func (ui *ModalUi) Bounds() (width, height int) {
	return config.Width, config.Height
}

func (ui *ModalUi) Draw() *ebiten.Image {
	// Draw background
	ui.screen.Fill(color.RGBA{0, 0, 0, 150})
	r, g, b, _ := bgColor.RGBA()
	ui.contentScreen.Fill(color.RGBA{uint8(r), uint8(g), uint8(b), 0})

	drawStackLayout(ui.contentScreen, ui.stackLayout)
	// Draw content onto modal body
	pos := ebiten.GeoM{}
	pos.Translate(float64(paddingX), float64((config.Height-ui.contentScreen.Bounds().Dy())/2))
	ui.screen.DrawImage(ui.contentScreen, &ebiten.DrawImageOptions{
		GeoM: pos,
	})

	return ui.screen
}
