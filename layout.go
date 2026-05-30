package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type SwitchLayout struct {
	interval int
	children []UiElement

	currentIndex    int
	frame           int
	transitionFrame int
	transition      bool
	image           *ebiten.Image
	curSnap         *ebiten.Image
	nextSnap        *ebiten.Image
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

	return config.Width, maxHeight
}

func (l *SwitchLayout) Init() {
	// Initialize all children
	for _, child := range l.children {
		child.Init()
	}

	_, height := l.Bounds()
	l.image = ebiten.NewImage(config.Width, height)
}

func (l *SwitchLayout) Draw() *ebiten.Image {
	l.image.Fill(bgColor)

	if l.transition {
		// Snapshot both children at the start of the slide so we don't pay
		// for their full Draw() on every one of the ~90 transition frames.
		if l.curSnap == nil {
			l.curSnap = l.children[l.currentIndex].Draw()
			l.nextSnap = l.children[(l.currentIndex+1)%len(l.children)].Draw()
		}

		pos := ebiten.GeoM{}
		pos.Translate(float64(l.transitionFrame), 0)
		l.image.DrawImage(l.curSnap, &ebiten.DrawImageOptions{GeoM: pos})

		pos.Reset()
		pos.Translate(float64(l.transitionFrame-config.Width), 0)
		l.image.DrawImage(l.nextSnap, &ebiten.DrawImageOptions{GeoM: pos})

		l.transitionFrame += 12

		if l.transitionFrame >= config.Width {
			l.currentIndex = (l.currentIndex + 1) % len(l.children)
			l.transition = false
			l.transitionFrame = 0
			l.curSnap = nil
			l.nextSnap = nil
		}
	} else {
		childImage := l.children[l.currentIndex].Draw()
		l.image.DrawImage(childImage, &ebiten.DrawImageOptions{})
	}

	if l.frame%l.interval == 0 {
		l.transition = true
	}
	l.frame++

	return l.image
}
