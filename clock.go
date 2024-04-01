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
	return WIDTH, fontHeight
}

func (ui *ClockUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	text.Draw(ui.screen, time.Now().Format("15:04"), defaultFont, 0, fontHeight, textColor)
	text.Draw(ui.screen, time.Now().In(ui.moscowLoc).Format("15:04"), defaultFont, 50, fontHeight, textColor)
	text.Draw(ui.screen, time.Now().In(ui.washingtonLoc).Format("15:04"), defaultFont, 100, fontHeight, textColor)

	return ui.screen
}
