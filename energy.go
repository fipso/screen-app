package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
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

var energyUsage map[string]int

func init() {
	energyUsage = make(map[string]int)
}

type EnergyUi struct {
	screen *ebiten.Image
}

func (ui *EnergyUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *EnergyUi) Bounds() (width, height int) {
	return config.Width, fontHeight+(fontHeight+linePadding) * len(config.Refoss_Energy_Devices)
}

func (ui *EnergyUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	// Draw the energy usage for each device
	for i, device := range config.Refoss_Energy_Devices {
		usage := energyUsage[device.Address]
		text.Draw(ui.screen, fmt.Sprintf("%s: %dW", strings.ToLower(device.Name), usage), defaultFont, 0, fontHeight+(linePadding+fontHeight)*i, textColor)
	}

	return ui.screen
}

func pollEnergyDevices() {
	for {
		for _, device := range config.Refoss_Energy_Devices {
			err := pollDevice(device.Address, device.UUID)
			if err != nil {
				log.Println("Error polling refoss device:", device.Address, err)
			}
		}

		time.Sleep(10 * time.Second)
	}
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

func pollDevice(address, uuid string) error {
	// Generate timestamp in seconds
	timestamp := time.Now().Unix()

	// Generate message ID based on the Java implementation
	messageId := generateMessageId()

	// Generate sign based on the Java implementation
	sign := generateSign(messageId, config.Refoss_Key, fmt.Sprintf("%d", timestamp))

	reqData := RefossDeviceConfigRequest{
		Header: RefossDeviceConfigHeader{
			Method:         "GET",
			From:           fmt.Sprintf("%s/config", address),
			MessageID:      messageId,
			PayloadVersion: 1,
			Namespace:      "Appliance.Control.Electricity",
			UUID:           uuid,
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/config", address), bytes.NewReader(reqDataJson))
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

	energyUsage[address] = resData.Payload.Electricity.Power / 1000

	return nil
}
