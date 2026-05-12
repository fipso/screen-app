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
	ui.moscowLoc, _ = time.LoadLocation("Europe/Moscow")
	ui.washingtonLoc, _ = time.LoadLocation("America/New_York")

	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *ClockUi) Bounds() (width, height int) {
	return config.Width, 156
}

func (ui *ClockUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	contentWidth := config.Width - 2*paddingX
	cellWidth := contentWidth / 3
	now := time.Now()
	cells := []struct {
		t    string
		city string
	}{
		{now.Format("15:04"), "ber"},
		{now.In(ui.moscowLoc).Format("15:04"), "mosc"},
		{now.In(ui.washingtonLoc).Format("15:04"), "wash"},
	}

	const timeBaseline = 100
	const cityBaseline = 148

	for i, c := range cells {
		x := i * cellWidth
		text.Draw(ui.screen, c.t, clockFont, x, timeBaseline, textColor)
		text.Draw(ui.screen, c.city, tinyFont, x, cityBaseline, dimColor)
	}

	return ui.screen
}
