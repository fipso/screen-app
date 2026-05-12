package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

var refossHTTPClient = &http.Client{Timeout: 5 * time.Second}

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

type RefossToggleRequest struct {
	Header  RefossDeviceConfigHeader `json:"header"`
	Payload RefossTogglePayload      `json:"payload"`
}

type RefossTogglePayload struct {
	Togglex RefossToggleX `json:"togglex"`
}

type RefossToggleX struct {
	Channel int `json:"channel"`
	Onoff   int `json:"onoff"`
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
	graph        *chart.Chart
	renderBuf    *bytes.Buffer
	rgbaBuf      *image.RGBA
}

type EnergySensorState struct {
	deviceConfig RefossEnergyDeviceConfig
	timestamps   []time.Time
	values       []float64
	series       *chart.TimeSeries
}

func (ui *EnergyUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
	ui.chartImage = ebiten.NewImage(width-50, 800)
	ui.renderBuf = bytes.NewBuffer(make([]byte, 0, 1024*1024))
	ui.rgbaBuf = image.NewRGBA(image.Rect(0, 0, width-50, 800))

	for _, device := range config.Energy.Devices {
		ui.deviceStates = append(ui.deviceStates, &EnergySensorState{
			deviceConfig: device,
			timestamps:   []time.Time{},
			values:       []float64{},
			series: &chart.TimeSeries{
				Name: device.Name,
				Style: chart.Style{
					StrokeWidth: 1.4,
				},
			},
		})
	}

	ui.initGraph()
	go ui.pollDeviceStates()
}

func (ui *EnergyUi) Bounds() (width, height int) {
	return config.Width, sectionHeaderHeight + 1180
}

func (ui *EnergyUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	usage := 0.0
	for _, device := range ui.deviceStates {
		if len(device.series.YValues) == 0 {
			continue
		}
		usage += device.series.YValues[len(device.series.YValues)-1]
	}

	contentY := drawSectionHeader(ui.screen, "power", fmt.Sprintf("%dw", int(usage)), 0)

	pos := ebiten.GeoM{}
	pos.Translate(0, float64(contentY+20))
	ui.screen.DrawImage(ui.chartImage, &ebiten.DrawImageOptions{
		GeoM: pos,
	})

	text.Draw(
		ui.screen,
		fmt.Sprintf("total consooomtion:\n\n   %dW %.2f€/h", int(usage), usage/1000*0.35),
		defaultFont,
		0,
		contentY+980,
		textColor,
	)

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
				time.Sleep(time.Second * 2)
			}
		}(device)
	}

	// Update chart less often than we poll — rendering ~300k points was the
	// CPU drag on the helper threads.
	go func() {
		for {
			time.Sleep(time.Second * 3)
			ui.updateGraph()
		}
	}()
}

// chartMaxPoints caps how many samples we hand to go-chart per series. With
// 6h × 0.5 Hz we have ~10k points per device; the chart canvas is ~1k pixels
// wide so anything beyond a few hundred is wasted work + GC pressure.
const chartMaxPoints = 600

func decimate(timestamps []time.Time, values []float64) ([]time.Time, []float64) {
	n := len(values)
	if n <= chartMaxPoints {
		return timestamps, values
	}
	stride := n / chartMaxPoints
	if stride < 1 {
		stride = 1
	}
	outLen := (n + stride - 1) / stride
	xs := make([]time.Time, 0, outLen)
	ys := make([]float64, 0, outLen)
	for i := 0; i < n; i += stride {
		xs = append(xs, timestamps[i])
		ys = append(ys, values[i])
	}
	// Always keep the most recent sample so the right edge of the chart
	// reflects the latest reading.
	if len(xs) == 0 || !xs[len(xs)-1].Equal(timestamps[n-1]) {
		xs = append(xs, timestamps[n-1])
		ys = append(ys, values[n-1])
	}
	return xs, ys
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
	if err != nil {
		return err
	}

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

	// log.Println(e.deviceConfig.Address)
	// spew.Dump(resData.Payload.Electricity.Power/1000)

	now := time.Now()
	p := float64(resData.Payload.Electricity.Power) / 1000
	e.values = append(e.values, p)
	e.timestamps = append(e.timestamps, now)

	// Get history length
	diff := now.Sub(e.timestamps[0])
	if diff > time.Hour*time.Duration(config.Energy.MaxHistoryHours) {
		// Drop oldest value
		e.values = e.values[1:]
		e.timestamps = e.timestamps[1:]
	}

	return nil
}

func (d *RefossEnergyDeviceConfig) SetPlugState(on bool) error {
	messageId := generateMessageId()
	timestamp := time.Now().Unix()

	key, ok := config.Energy.Profiles[d.Profile]
	if !ok {
		return fmt.Errorf("meross profile %s not found", d.Profile)
	}

	sign := generateSign(messageId, key, fmt.Sprintf("%d", timestamp))

	url := fmt.Sprintf("%s/config", d.Address)

	onoff := 0
	if on {
		onoff = 1
	}

	reqData := RefossToggleRequest{
		Header: RefossDeviceConfigHeader{
			Method:         "SET",
			From:           url,
			MessageID:      messageId,
			PayloadVersion: 1,
			Namespace:      "Appliance.Control.ToggleX",
			UUID:           d.UUID,
			Sign:           sign,
			TriggerSrc:     "GoClient",
			Timestamp:      int(timestamp),
		},
		Payload: RefossTogglePayload{
			Togglex: RefossToggleX{
				Channel: 0,
				Onoff:   onoff,
			},
		},
	}

	reqDataJson, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	const maxAttempts = 5
	backoff := time.Second
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest("POST", url, bytes.NewReader(reqDataJson))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Proxy-Connection", "keep-alive")
		req.Header.Set("User-Agent", "intellect_socket/1.10.0 (iPhone; iOS 18.3.2; Scale/3.00)")
		req.Header.Set("Accept-Language", "en-DE;q=1, de-DE;q=0.9")

		resp, err := refossHTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("plug %s set state request failed: %w", d.UUID, err)
		} else if resp.StatusCode > 399 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("plug %s set state failed: %d: %s", d.UUID, resp.StatusCode, string(body))
		} else {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			return nil
		}

		if attempt < maxAttempts {
			log.Printf("plug %s set state attempt %d/%d failed: %v; retrying in %s", d.UUID, attempt, maxAttempts, lastErr, backoff)
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return lastErr
}

func (ui *EnergyUi) initGraph() {
	fill := drawing.Color{R: bgColor.R, G: bgColor.G, B: bgColor.B, A: bgColor.A}
	font := drawing.Color{R: textColor.R, G: textColor.G, B: textColor.B, A: textColor.A}
	width, _ := ui.Bounds()

	ui.graph = &chart.Chart{
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
				Max: 800,
			},
		},
		Series: make([]chart.Series, len(ui.deviceStates)),
		Width:  width - 50,
		Height: 800,
	}

	// Add all series to the chart
	for i, device := range ui.deviceStates {
		ui.graph.Series[i] = device.series
	}

	ui.graph.Elements = append(ui.graph.Elements, chart.LegendLeft(ui.graph, chart.Style{
		FillColor:   fill,
		FontColor:   font,
		StrokeColor: font,
		FontSize:    14,
	}))
}

func (ui *EnergyUi) updateGraph() {
	// Update colors for day/night mode
	fill := drawing.Color{R: bgColor.R, G: bgColor.G, B: bgColor.B, A: bgColor.A}
	font := drawing.Color{R: textColor.R, G: textColor.G, B: textColor.B, A: textColor.A}
	ui.graph.Canvas.FillColor = fill
	ui.graph.YAxis.Style.FontColor = font

	// Recreate legend with updated colors
	ui.graph.Elements = []chart.Renderable{chart.LegendLeft(ui.graph, chart.Style{
		FillColor:   fill,
		FontColor:   font,
		StrokeColor: font,
		FontSize:    14,
	})}

	// Update series data pointers
	for _, device := range ui.deviceStates {
		if len(device.deviceConfig.Aggregate) == 0 {
			xs, ys := decimate(device.timestamps, device.values)
			device.series.XValues = xs
			device.series.YValues = ys
			continue
		}

		// Apply aggregation
		aggregatedValues := make([]float64, len(device.values))

		for i, v := range device.values {
			t := device.timestamps[i]

			newValue := v
			for _, aggr := range device.deviceConfig.Aggregate {
				// Find other device by uuid
				var otherDevice *EnergySensorState
				for _, d := range ui.deviceStates {
					if d.deviceConfig.UUID == aggr.Device {
						otherDevice = d
						break
					}
				}
				if otherDevice == nil {
					// Other device not found skip aggregation task
					continue
				}

				// Find latest value at or before timestamp of other device
				otherDeviceValue := 0.0
				for j := len(otherDevice.timestamps) - 1; j >= 0; j-- {
					if !otherDevice.timestamps[j].After(t) {
						otherDeviceValue = otherDevice.values[j]
						break
					}
				}

				switch aggr.Operation {
				case AggrOpAdd:
					newValue += otherDeviceValue
				case AggrOpSub:
					newValue -= otherDeviceValue
				}
			}

			aggregatedValues[i] = newValue
		}

		xs, ys := decimate(device.timestamps, aggregatedValues)
		device.series.XValues = xs
		device.series.YValues = ys
	}

	// Reuse buffer
	ui.renderBuf.Reset()
	err := ui.graph.Render(chart.PNG, ui.renderBuf)
	if err != nil {
		log.Println("Could not render graph:", err)
		return
	}

	img, err := png.Decode(ui.renderBuf)
	if err != nil {
		log.Println("Could not decode graph png:", err)
		return
	}

	// Convert to RGBA reusing buffer and write pixels directly
	bounds := img.Bounds()
	draw.Draw(ui.rgbaBuf, bounds, img, bounds.Min, draw.Src)
	ui.chartImage.WritePixels(ui.rgbaBuf.Pix)
}
