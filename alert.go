package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type AlertUi struct {
	screen *ebiten.Image
	msg    string
}

func (ui *AlertUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *AlertUi) Bounds() (width, height int) {
	return config.Width - paddingX, 200
}

func (ui *AlertUi) Draw() *ebiten.Image {
	w, _ := ui.Bounds()

	// Draw background with some transparency
	r, b, g, _ := bgColor.RGBA()
	ui.screen.Fill(color.RGBA{uint8(r), uint8(g), uint8(b), 220})

	// Draw rectangle border

	text.Draw(ui.screen, string(""), faFont, w/2, 50, textColor)
	text.Draw(ui.screen, ui.msg, defaultFont, 0, 150, textColor)

	return ui.screen
}
