package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type SwitchLayout struct {
	interval int
	children []UiElement

	currentIndex    int
	frame           int
	transitionFrame int
	transition      bool
	image           *ebiten.Image
}

func (l *SwitchLayout) Bounds() (width, height int) {
	// Find child with max height
	maxHeight := 0
	for _, child := range l.children {
		_, h := child.Bounds()
		if h > maxHeight {
			maxHeight = h
		}
	}

	return WIDTH, maxHeight
}

func (l *SwitchLayout) Init() {
	// Initialize all children
	for _, child := range l.children {
		child.Init()
	}

	_, height := l.Bounds()
	l.image = ebiten.NewImage(WIDTH, height)
}

func (l *SwitchLayout) Draw() *ebiten.Image {
	l.image.Fill(bgColor)

	child := l.children[l.currentIndex]
	childImage := child.Draw()

	if l.transition {
		nextChild := l.children[(l.currentIndex+1)%len(l.children)]
		nextChildImage := nextChild.Draw()

		pos := ebiten.GeoM{}

		// Draw the current child
		pos.Translate(float64(l.transitionFrame), 0)
		colorm.DrawImage(l.image, childImage, colorm.ColorM{}, &colorm.DrawImageOptions{
			GeoM: pos,
		})

		// Draw the next child
		pos.Reset()
		pos.Translate(float64(l.transitionFrame-WIDTH), 0)
		colorm.DrawImage(l.image, nextChildImage, colorm.ColorM{}, &colorm.DrawImageOptions{
			GeoM: pos,
		})

		l.transitionFrame+=12

		if l.transitionFrame == WIDTH {
			l.currentIndex = (l.currentIndex + 1) % len(l.children)
			l.transition = false
			l.transitionFrame = 0
		}
	} else {
		l.image.DrawImage(childImage, &ebiten.DrawImageOptions{})
	}

	if l.frame%l.interval == 0 {
		l.transition = true
	}
	l.frame++

	return l.image
}
