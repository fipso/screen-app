package main

import (
	"fmt"
	"image/color"
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

type RoomUi struct {
	screen    *ebiten.Image
	things    []*roomThingState
	labelFont font.Face
	valueFont font.Face
}

type roomThingState struct {
	cfg                      RoomThing
	energy                   *EnergySensorState
	temp, rh, extra          float64
	hasTemp, hasRh, hasExtra bool
}

var (
	roomColorOn      = color.RGBA{60, 180, 75, 255}
	roomColorOff     = color.RGBA{60, 60, 60, 255}
	roomColorUnknown = color.RGBA{170, 170, 170, 255}
)

func (ui *RoomUi) Init() {
	w, h := ui.Bounds()
	ui.screen = ebiten.NewImage(w, h)

	ui.labelFont = loadFont("assets/fonts/OpenSans-Regular.ttf", float64(config.Room.LabelFontSize))
	ui.valueFont = loadFont("assets/fonts/OpenSans-Regular.ttf", float64(config.Room.ValueFontSize))

	for i := range config.Room.Things {
		cfg := config.Room.Things[i]
		st := &roomThingState{cfg: cfg}
		if cfg.RefossUUID != "" {
			if es, ok := getEnergyState(cfg.RefossUUID); ok {
				st.energy = es
			} else {
				log.Printf("RoomUi: refoss UUID %s for %s not in config.Energy.Devices", cfg.RefossUUID, cfg.Label)
			}
		}
		ui.things = append(ui.things, st)
	}

	go func() {
		mqttService.WaitReady()
		for _, st := range ui.things {
			st := st
			if st.cfg.MqttTemp != "" {
				mqttService.On(st.cfg.MqttTemp, func(_ mqtt.Client, m mqtt.Message) {
					if v, err := strconv.ParseFloat(string(m.Payload()), 64); err == nil {
						st.temp = v
						st.hasTemp = true
					}
				})
			}
			if st.cfg.MqttRh != "" {
				mqttService.On(st.cfg.MqttRh, func(_ mqtt.Client, m mqtt.Message) {
					if v, err := strconv.ParseFloat(string(m.Payload()), 64); err == nil {
						st.rh = v
						st.hasRh = true
					}
				})
			}
			if st.cfg.MqttExtra != "" {
				mqttService.On(st.cfg.MqttExtra, func(_ mqtt.Client, m mqtt.Message) {
					if v, err := strconv.ParseFloat(string(m.Payload()), 64); err == nil {
						st.extra = v
						st.hasExtra = true
					}
				})
			}
		}
	}()
}

func (ui *RoomUi) Bounds() (int, int) {
	w := config.Room.Width
	if w == 0 {
		w = config.Width
	}
	h := config.Room.Height
	if h == 0 {
		h = 800
	}
	return w, h
}

func (ui *RoomUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	for _, st := range ui.things {
		col := stateColor(st)
		ui.drawShape(st, col)
		ui.drawLabel(st)
	}

	return ui.screen
}

func stateColor(st *roomThingState) color.RGBA {
	if st.cfg.RefossUUID == "" || st.energy == nil {
		return roomColorUnknown
	}
	if on, ok := st.energy.togglex[st.cfg.RefossChannel]; ok {
		if on {
			return roomColorOn
		}
		return roomColorOff
	}
	if len(st.energy.values) == 0 {
		return roomColorUnknown
	}
	last := st.energy.values[len(st.energy.values)-1]
	if last > 1 { // values are in W; treat <=1W as off (idle draw / noise)
		return roomColorOn
	}
	return roomColorOff
}

func (ui *RoomUi) drawShape(st *roomThingState, col color.RGBA) {
	switch st.cfg.Shape {
	case RoomShapeRect:
		vector.StrokeRect(
			ui.screen,
			float32(st.cfg.X), float32(st.cfg.Y),
			float32(st.cfg.W), float32(st.cfg.H),
			2, col, true,
		)
	case RoomShapeCircle:
		vector.StrokeCircle(
			ui.screen,
			float32(st.cfg.X), float32(st.cfg.Y),
			float32(st.cfg.Radius),
			2, col, true,
		)
	}
}

type renderLine struct {
	text    string
	font    font.Face
	h       int
	isLabel bool
}

func (ui *RoomUi) drawLabel(st *roomThingState) {
	label := st.cfg.Label

	var energyText string
	if st.energy != nil && !st.cfg.HideEnergy && len(st.energy.values) > 0 {
		energyText = fmt.Sprintf("%.0fW", st.energy.values[len(st.energy.values)-1])
	}

	var otherLines []string
	if st.hasTemp {
		otherLines = append(otherLines, fmt.Sprintf("%.1f°", st.temp))
	}
	if st.hasRh {
		otherLines = append(otherLines, fmt.Sprintf("%.0f%%", st.rh))
	}
	if st.hasExtra {
		l := st.cfg.ExtraLabel
		if l == "" {
			l = "extra"
		}
		otherLines = append(otherLines, fmt.Sprintf("%s %.0f", l, st.extra))
	}

	anchor := st.cfg.EnergyAnchor
	labelLineH := lineHeight(ui.labelFont)
	valueLineH := lineHeight(ui.valueFont)

	// Build the vertical stack. Energy joins the stack only when anchored
	// top/bottom; left/right take it out and place it adjacent to the label.
	var stack []renderLine
	if anchor == EnergyAnchorTop && energyText != "" {
		stack = append(stack, renderLine{energyText, ui.valueFont, valueLineH, false})
	}
	if label != "" {
		stack = append(stack, renderLine{label, ui.labelFont, labelLineH, true})
	}
	if anchor == EnergyAnchorBottom && energyText != "" {
		stack = append(stack, renderLine{energyText, ui.valueFont, valueLineH, false})
	}
	for _, l := range otherLines {
		stack = append(stack, renderLine{l, ui.valueFont, valueLineH, false})
	}

	sideEnergy := energyText != "" && (anchor == EnergyAnchorLeft || anchor == EnergyAnchorRight)

	if len(stack) == 0 && !sideEnergy {
		return
	}

	totalH := 0
	for _, l := range stack {
		totalH += l.h
	}

	var labelDrawnX, labelDrawnY, labelDrawnW int
	labelDrawn := false

	if st.cfg.Shape == RoomShapeNone {
		// Text-only: top-left anchor at (X, Y), left-aligned.
		ty := st.cfg.Y
		for _, l := range stack {
			ty += l.h
			text.Draw(ui.screen, l.text, l.font, st.cfg.X, ty, textColor)
			if l.isLabel {
				labelDrawnX = st.cfg.X
				labelDrawnY = ty
				labelDrawnW = text.BoundString(l.font, l.text).Dx()
				labelDrawn = true
			}
		}
	} else if len(stack) > 0 {
		cx, cy := thingCenter(st.cfg)
		ty := cy - totalH/2 + stack[0].h
		for i, l := range stack {
			bounds := text.BoundString(l.font, l.text)
			tx := cx - bounds.Dx()/2
			text.Draw(ui.screen, l.text, l.font, tx, ty, textColor)
			if l.isLabel {
				labelDrawnX = tx
				labelDrawnY = ty
				labelDrawnW = bounds.Dx()
				labelDrawn = true
			}
			if i+1 < len(stack) {
				ty += stack[i+1].h
			}
		}
	}

	if sideEnergy {
		bounds := text.BoundString(ui.valueFont, energyText)
		const pad = 10
		var ex, ey int
		if labelDrawn {
			ey = labelDrawnY
			if anchor == EnergyAnchorLeft {
				ex = labelDrawnX - pad - bounds.Dx()
			} else {
				ex = labelDrawnX + labelDrawnW + pad
			}
		} else if st.cfg.Shape != RoomShapeNone {
			cx, cy := thingCenter(st.cfg)
			ey = cy + valueLineH/3
			ex = cx - bounds.Dx()/2
		} else {
			ey = st.cfg.Y + valueLineH
			ex = st.cfg.X
		}
		text.Draw(ui.screen, energyText, ui.valueFont, ex, ey, textColor)
	}
}

func thingCenter(cfg RoomThing) (int, int) {
	switch cfg.Shape {
	case RoomShapeRect:
		return cfg.X + cfg.W/2, cfg.Y + cfg.H/2
	case RoomShapeCircle:
		return cfg.X, cfg.Y
	default:
		return cfg.X, cfg.Y
	}
}

func lineHeight(face font.Face) int {
	m := face.Metrics()
	return (m.Ascent + m.Descent).Ceil() + 2
}
