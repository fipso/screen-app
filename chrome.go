package main

import (
	"unicode"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const sectionHeaderHeight = 68

// drawSectionHeader renders an industrial-style section banner at the top of
// img, starting at y. Returns the y where content should begin below the rule.
//
// Layout (origin at the widget's local 0,0; the widget itself is translated by
// paddingX when composited, so we leave another paddingX of breathing room on
// the right edge to mirror the left margin):
//
//   ● label                                                            right
//   ───────────────────────────────────────────────────────────────────────
func drawSectionHeader(img *ebiten.Image, label, right string, y int) int {
	contentWidth := img.Bounds().Dx() - 2*paddingX
	const baseline = 36

	if r, size := utf8.DecodeRuneInString(label); r != utf8.RuneError {
		label = string(unicode.ToUpper(r)) + label[size:]
	}

	vector.DrawFilledCircle(img, 8, float32(y+baseline-9), 7, accentColor, true)
	text.Draw(img, label, smallBoldFont, 28, y+baseline, textColor)

	if right != "" {
		b := text.BoundString(smallFont, right)
		text.Draw(img, right, smallFont, contentWidth-b.Dx(), y+baseline, dimColor)
	}

	ruleY := y + baseline + 16
	vector.StrokeLine(img, 0, float32(ruleY), float32(contentWidth), float32(ruleY), 1, ruleColor, true)
	return ruleY + 16
}
