package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type AlertUi struct {
	screen *ebiten.Image
	msg    string
	// icon is a FontAwesome glyph (e.g. ""). Empty falls back to the
	// warning triangle used for the original doorbell modal.
	icon string
}

func (ui *AlertUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *AlertUi) Bounds() (width, height int) {
	return config.Width - paddingX*2, 400
}

func (ui *AlertUi) Draw() *ebiten.Image {
	w, _ := ui.Bounds()

	// Draw background with some transparency
	r, b, g, _ := bgColor.RGBA()
	ui.screen.Fill(color.RGBA{uint8(r), uint8(g), uint8(b), 220})

	icon := ui.icon
	if icon == "" {
		icon = "" // FA triangle-exclamation
	}
	text.Draw(ui.screen, icon, faFont, w/2-48*2, 48*2, textColor)
	if ui.msg != "" {
		text.Draw(ui.screen, ui.msg, defaultFont, 0, 200, textColor)
	}

	return ui.screen
}
