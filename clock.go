package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type ClockUi struct {
	screen *ebiten.Image

	moscowLoc     *time.Location
	washingtonLoc *time.Location
}

func (ui *ClockUi) Init() {
	// Load locations
	ui.moscowLoc, _ = time.LoadLocation("Europe/Moscow")
	ui.washingtonLoc, _ = time.LoadLocation("America/New_York")

	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *ClockUi) Bounds() (width, height int) {
	return config.Width, fontHeight + linePadding*4
}

func (ui *ClockUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	text.Draw(ui.screen, time.Now().Format("15:04"), clockFont, fontWidth, fontHeight+linePadding*2, textColor)
	text.Draw(ui.screen, "BER", tinyFont, fontWidth*4, fontHeight+linePadding*2, textColor)

	text.Draw(ui.screen, time.Now().In(ui.moscowLoc).Format("15:04"), clockFont, fontWidth*6, fontHeight+linePadding*2, textColor)
	text.Draw(ui.screen, "MOSC", tinyFont, fontWidth*9, fontHeight+linePadding*2, textColor)

	text.Draw(ui.screen, time.Now().In(ui.washingtonLoc).Format("15:04"), clockFont, fontWidth*11, fontHeight+linePadding*2, textColor)
	text.Draw(ui.screen, "WASH", tinyFont, fontWidth*14+fontWidth/2, fontHeight+linePadding*2, textColor)

	return ui.screen
}
