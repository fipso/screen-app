package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type RefossDeviceConfigHeader struct {
	MessageID      string `json:"messageId"`
	Method         string `json:"method"`
	From           string `json:"from"`
	PayloadVersion int    `json:"payloadVersion"`
	Namespace      string `json:"namespace"`
	UUID           string `json:"uuid"`
	Sign           string `json:"sign"`
	TriggerSrc     string `json:"triggerSrc"`
	Timestamp      int    `json:"timestamp"`
}

type RefossDeviceConfigRequest struct {
	Header  RefossDeviceConfigHeader `json:"header"`
	Payload struct {
		Electricity RefossDeviceConfigRequestElectricity `json:"electricity"`
	} `json:"payload"`
}

type RefossDeviceConfigRequestElectricity struct {
	Channel int `json:"channel"`
}

type RefossDeviceConfigResponse struct {
	Header  RefossDeviceConfigHeader `json:"header"`
	Payload struct {
		Electricity struct {
			Channel int `json:"channel"`
			Current int `json:"current"`
			Voltage int `json:"voltage"`
			Power   int `json:"power"`
			Config  struct {
				VoltageRatio          int `json:"voltageRatio"`
				ElectricityRatio      int `json:"electricityRatio"`
				MaxElectricityCurrent int `json:"maxElectricityCurrent"`
				PowerRatio            int `json:"powerRatio"`
			} `json:"config"`
		} `json:"electricity"`
	} `json:"payload"`
}

type EnergyUi struct {
	screen *ebiten.Image

	deviceStates []*EnergySensorState
	chartImage   *ebiten.Image
}

type EnergySensorState struct {
	lock         sync.Mutex
	deviceConfig RefossEnergyDeviceConfig
	timestamps   []time.Time
	values       []float64
}

func (ui *EnergyUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
	ui.chartImage = ebiten.NewImage(width, 600)

	for _, device := range config.Energy.Devices {
		ui.deviceStates = append(ui.deviceStates, &EnergySensorState{
			deviceConfig: device,
			timestamps:   []time.Time{},
			values:       []float64{},
		})
	}

	go ui.pollDeviceStates()
}

func (ui *EnergyUi) Bounds() (width, height int) {
	return config.Width, 1300
}

func (ui *EnergyUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	// Draw the energy usage for each device
	// for i, device := range config.Refoss_Energy_Devices {
	// 	usage := energyUsage[device.Address]
	// 	text.Draw(ui.screen, fmt.Sprintf("%s: %dW", strings.ToLower(device.Name), usage), defaultFont, 0, fontHeight+(linePadding+fontHeight)*i, textColor)
	// }

	text.Draw(ui.screen, "power", defaultFont, 0, 100, textColor)

	pos := ebiten.GeoM{}
	pos.Translate(0, 140)
	ui.screen.DrawImage(ui.chartImage, &ebiten.DrawImageOptions{
		GeoM: pos,
	})

	usage := 0.0
	for _, device := range ui.deviceStates {
		if len(device.values) == 0 {
			continue
		}
		usage += device.values[len(device.values)-1]
	}
	text.Draw(ui.screen, fmt.Sprintf("total consooomtion:\n\n   %dW %.2f€/h", int(usage), usage/1000*0.35), defaultFont, 0, 1100, textColor)

	return ui.screen
}

func (ui *EnergyUi) pollDeviceStates() {
	for _, device := range ui.deviceStates {
		go func(s *EnergySensorState) {
			for {
				err := s.fetchState()
				if err != nil {
					log.Println("Error polling refoss device:", s.deviceConfig.Address, err)
				}
				time.Sleep(time.Second)
			}
		}(device)
	}

	// Update chart at 1 fps
	go func() {
		for {
			time.Sleep(time.Second * 1)
			ui.updateGraph()
		}
	}()
}

// generateRandomString creates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// md5Hash generates an MD5 hash of the input string
func md5Hash(input string) string {
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// generateMessageId replicates the g3() function from Java
func generateMessageId() string {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Generate random string of 16 chars (k(16))
	randomStr := generateRandomString(16)

	// Get current timestamp in seconds (m())
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Combine and hash (g())
	return md5Hash(randomStr + timestamp)
}

// generateSign replicates the h3() function from Java
func generateSign(messageId, key, timestamp string) string {
	// Concatenate the three parameters and hash
	return md5Hash(messageId + key + timestamp)
}

func (e *EnergySensorState) fetchState() error {
	// Generate message ID based on the Java implementation
	messageId := generateMessageId()

	timestamp := time.Now().Unix()

	// Get key for profile
	key, ok := config.Energy.Profiles[e.deviceConfig.Profile]
	if !ok {
		return fmt.Errorf("meross profile %s not found", e.deviceConfig.Profile)
	}

	// Generate sign based on the Java implementation
	sign := generateSign(messageId, key, fmt.Sprintf("%d", timestamp))

	url := fmt.Sprintf("%s/config", e.deviceConfig.Address)

	reqData := RefossDeviceConfigRequest{
		Header: RefossDeviceConfigHeader{
			Method:         "GET",
			From:           url,
			MessageID:      messageId,
			PayloadVersion: 1,
			Namespace:      "Appliance.Control.Electricity",
			UUID:           e.deviceConfig.UUID,
			Sign:           sign,
			TriggerSrc:     "GoClient",
			Timestamp:      int(timestamp),
		},
		Payload: struct {
			Electricity RefossDeviceConfigRequestElectricity "json:\"electricity\""
		}{
			Electricity: RefossDeviceConfigRequestElectricity{
				Channel: 0,
			},
		},
	}
	reqDataJson, err := json.Marshal(reqData)

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqDataJson))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Proxy-Connection", "keep-alive")
	req.Header.Set("User-Agent", "intellect_socket/1.10.0 (iPhone; iOS 18.3.2; Scale/3.00)")
	req.Header.Set("Accept-Language", "en-DE;q=1, de-DE;q=0.9")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var resData RefossDeviceConfigResponse
	err = json.Unmarshal(body, &resData)
	if err != nil {
		return err
	}

	e.lock.Lock()
	e.values = append(e.values, float64(resData.Payload.Electricity.Power)/1000)
	e.timestamps = append(e.timestamps, time.Now())
	e.lock.Unlock()

	return nil
}

func (ui *EnergyUi) newGraph() *chart.Chart {
	// Find highest power usage
	highest := 0.0
	for _, device := range ui.deviceStates {
		if len(device.values) == 0 {
			continue
		}
		v := device.values[len(device.values)-1]
		if v > highest {
			highest = v
		}
	}

	fill := drawing.Color{R: bgColor.R, G: bgColor.G, B: bgColor.B, A: bgColor.A}
	font := drawing.Color{R: textColor.R, G: textColor.G, B: textColor.B, A: textColor.A}
	width, _ := ui.Bounds()
	graph := chart.Chart{
		DPI: 100,
		Background: chart.Style{
			FillColor: chart.ColorTransparent,
		},
		Canvas: chart.Style{
			FillColor: fill,
		},
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeValueFormatterWithFormat("15:04"),
		},
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontColor: font,
			},
			Range: &chart.ContinuousRange{
				Max: float64(int(highest/100))*100.0 + 150,
			},
		},
		Series: []chart.Series{},
		Width:  width - 50,
		Height: 800,
	}

	graph.Elements = append(graph.Elements, chart.LegendLeft(&graph, chart.Style{
		FillColor:   fill,
		FontColor:   font,
		StrokeColor: font,
		FontSize:    14,
	}))

	for _, device := range ui.deviceStates {
		if len(device.values) == 0 {
			continue
		}

		graph.Series = append(graph.Series, chart.TimeSeries{
			Name:    device.deviceConfig.Name,
			XValues: device.timestamps,
			YValues: device.values,
		})

		/*
		graph.Series = append(graph.Series, chart.AnnotationSeries{
			Annotations: []chart.Value2{
				{XValue: 600.0, YValue: 600.0, Label: "One"},
				{XValue: 2.0, YValue: 2.0, Label: "Two"},
				{XValue: 3.0, YValue: 3.0, Label: "Three"},
				{XValue: 4.0, YValue: 4.0, Label: "Four"},
				{XValue: 5.0, YValue: 5.0, Label: "Five"},
			},
			Style: chart.Style{
				FillColor: chart.ColorWhite,
			},
		})*/
	}

	return &graph
}

func (ui *EnergyUi) updateGraph() {
	graph := ui.newGraph()

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		log.Println("Could not render graph")
		log.Println(err)
		return
	}

	img, err := png.Decode(buffer)
	if err != nil {
		log.Println("Could not decode graph png")
		return
	}

	ui.chartImage = ebiten.NewImageFromImage(img)
}
