package main

import (
	"image/color"
	"sync"

	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// VPDChart represents a VPD chart that can be updated and drawn
type VPDChart struct {
	image       *ebiten.Image
	width       int
	height      int
	sensors     []SensorData
	sensorNames []string
	lock        sync.Mutex
}

// NewVPDChart creates a new VPD chart
func NewVPDChart(width, height int, sensors []SensorData, sensorNames []string) *VPDChart {
	//sensorsData := make([]SensorData, len(sensors))
	//copy(sensorsData, sensors)
	return &VPDChart{
		width:       width,
		height:      height,
		image:       ebiten.NewImage(width, height),
		sensors:     sensors,
		sensorNames: sensorNames,
	}
}

// Draw draws the VPD chart to the provided image
func (v *VPDChart) Update() {
	v.lock.Lock()
	defer v.lock.Unlock()

	// Clear the image with white background
	v.image.Fill(color.White)

	// Draw grid and scales
	// v.drawScalesAndGrid()

	// Draw VPD zones
	v.drawVPDZones()

	// Draw markers
	for i, sensor := range v.sensors {
		if sensor.humidLast == 0 || sensor.tempLast == 0 {
			continue
		}

		v.drawMarker(v.sensorNames[i], sensor.tempLast, sensor.humidLast, color.RGBA{0, 0, 0, 255})
	}
}

// drawVPDZones draws the VPD zones with appropriate colors
func (v *VPDChart) drawVPDZones() {
	for y := 0; y < v.height; y++ {
		for x := 0; x < v.width; x++ {
			temp := 13 + float64(y)*(26-13)/float64(v.height)
			rh := 81 - float64(x)*(81-19)/float64(v.width)
			vpd := calculateVPD(temp, rh)

			var c color.RGBA
			switch {
			case vpd < 0.6:
				c = color.RGBA{0, 0, 128, 100}
			case vpd < 1.0:
				c = color.RGBA{0, 128, 255, 100}
			case vpd < 1.4:
				c = color.RGBA{0, 200, 0, 100}
			case vpd < 1.8:
				c = color.RGBA{255, 255, 0, 100}
			default:
				c = color.RGBA{255, 0, 0, 100}
			}

			// Draw a single pixel with alpha blending
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			//op.ColorM.Scale(float64(c.R)/255, float64(c.G)/255, float64(c.B)/255, float64(c.A)/255)

			v.image.Set(x, y, c)

			// Create a 1x1 pixel image with the color
			// pixel := ebiten.NewImage(1, 1)
			// pixel.Fill(color.White)
			// v.image.DrawImage(pixel, op)
		}
	}
}

func (v *VPDChart) drawMarker(name string, temp, rh float64, col color.Color) {
	x := int((81 - rh) * float64(v.width) / (81 - 19))
	y := int((temp - 13) * float64(v.height) / (26 - 13))

	// Draw a filled circle
	vector.DrawFilledCircle(v.image, float32(x), float32(y), 8, col, true)
	text.Draw(v.image, name, smallFont, x+15, y+9, color.Black)
}

func (v *VPDChart) drawScalesAndGrid() {
	// Y-axis Temp
	for t := 13; t <= 26; t += 3 {
		y := int(float64(t-13) * float64(v.height) / float64(26-13))
		v.drawHorizontalLine(y, color.Gray{200})
		label := strconv.Itoa(t) + "°C"
		text.Draw(v.image, label, defaultFont, 5, y+5, textColor)
	}

	// X-axis RH
	for i, rh := range []int{81, 72, 63, 54, 46, 37, 28, 19} {
		x := int(float64(i) * float64(v.width) / 7)
		v.drawVerticalLine(x, color.Gray{200})
		label := strconv.Itoa(rh) + "%"
		text.Draw(v.image, label, defaultFont, x-10, 15, textColor)
	}
}

func (v *VPDChart) drawHorizontalLine(y int, c color.Color) {
	vector.StrokeLine(v.image, 0, float32(y), float32(v.width), float32(y), 1, c, true)
}

func (v *VPDChart) drawVerticalLine(x int, c color.Color) {
	vector.StrokeLine(v.image, float32(x), 0, float32(x), float32(v.height), 1, c, true)
}

func calculateVPD(temp float64, rh float64) float64 {
	es := 0.6108 * math.Exp((17.27*temp)/(temp+237.3))
	ea := es * rh / 100.0
	return es - ea
}
