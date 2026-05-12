package main

import (
	"image/color"
	"strconv"
	"strings"
	"time"
)

func applyDayNightPalette() {
	var p ThemePalette
	if h := time.Now().Hour(); h > 17 || h < 8 {
		p = config.Theme.Night
	} else {
		p = config.Theme.Day
	}
	bgColor = parseHex(p.Background)
	textColor = parseHex(p.Primary)
	dimColor = parseHex(p.Secondary)
	ruleColor = parseHex(p.Rule)
}

// parseHex accepts "#RRGGBB" or "#RRGGBBAA" (the leading # is optional).
// Defaults to opaque black on malformed input; in practice applyThemeDefaults
// fills every key first so this branch is unreachable.
func parseHex(s string) color.RGBA {
	s = strings.TrimPrefix(s, "#")
	var r, g, b uint64
	a := uint64(255)
	if len(s) >= 6 {
		r, _ = strconv.ParseUint(s[0:2], 16, 8)
		g, _ = strconv.ParseUint(s[2:4], 16, 8)
		b, _ = strconv.ParseUint(s[4:6], 16, 8)
		if len(s) == 8 {
			a, _ = strconv.ParseUint(s[6:8], 16, 8)
		}
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

func applyThemeDefaults(t *ThemeConfig) {
	fill := func(dst *string, def string) {
		if *dst == "" {
			*dst = def
		}
	}
	fill(&t.Night.Background, "#000000")
	fill(&t.Night.Primary, "#FFFFFF")
	fill(&t.Night.Secondary, "#FFFFFF6B")
	fill(&t.Night.Rule, "#FFFFFF29")
	fill(&t.Day.Background, "#F5F5F5")
	fill(&t.Day.Primary, "#000000")
	fill(&t.Day.Secondary, "#0000006B")
	fill(&t.Day.Rule, "#0000002E")
	fill(&t.Accent, "#14C814")
	fill(&t.Positive, "#14C814")
	fill(&t.Negative, "#FF0000")
}
