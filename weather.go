package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

type BrightskyPredictionRes struct {
	Weather []struct {
		Timestamp                  time.Time `json:"timestamp"`
		SourceID                   int       `json:"source_id"`
		Precipitation              float64   `json:"precipitation"`
		PressureMsl                float64   `json:"pressure_msl"`
		Sunshine                   float64   `json:"sunshine"`
		Temperature                float64   `json:"temperature"`
		WindDirection              int       `json:"wind_direction"`
		WindSpeed                  float64   `json:"wind_speed"`
		CloudCover                 int       `json:"cloud_cover"`
		DewPoint                   float64   `json:"dew_point"`
		RelativeHumidity           int       `json:"relative_humidity"`
		Visibility                 int       `json:"visibility"`
		WindGustDirection          any       `json:"wind_gust_direction"`
		WindGustSpeed              float64   `json:"wind_gust_speed"`
		Condition                  string    `json:"condition"`
		PrecipitationProbability   any       `json:"precipitation_probability"`
		PrecipitationProbability6H any       `json:"precipitation_probability_6h"`
		Solar                      float64   `json:"solar"`
		FallbackSourceIds          struct {
			Solar         int `json:"solar"`
			Sunshine      int `json:"sunshine"`
			Condition     int `json:"condition"`
			PressureMsl   int `json:"pressure_msl"`
			WindGustSpeed int `json:"wind_gust_speed"`
			Visibility    int `json:"visibility"`
			WindDirection int `json:"wind_direction"`
			WindSpeed     int `json:"wind_speed"`
			CloudCover    int `json:"cloud_cover"`
		} `json:"fallback_source_ids,omitempty"`
		Icon string `json:"icon"`
	} `json:"weather"`
	Sources []struct {
		ID              int       `json:"id"`
		DwdStationID    string    `json:"dwd_station_id"`
		ObservationType string    `json:"observation_type"`
		Lat             float64   `json:"lat"`
		Lon             float64   `json:"lon"`
		Height          float64   `json:"height"`
		StationName     string    `json:"station_name"`
		WmoStationID    string    `json:"wmo_station_id"`
		FirstRecord     time.Time `json:"first_record"`
		LastRecord      time.Time `json:"last_record"`
		Distance        float64   `json:"distance"`
	} `json:"sources"`
}

type BrightskyCurrentRes struct {
	Weather struct {
		SourceID            int       `json:"source_id"`
		Timestamp           time.Time `json:"timestamp"`
		CloudCover          int       `json:"cloud_cover"`
		Condition           string    `json:"condition"`
		DewPoint            float64   `json:"dew_point"`
		Solar10             float64   `json:"solar_10"`
		Solar30             float64   `json:"solar_30"`
		Solar60             float64   `json:"solar_60"`
		Precipitation10     float64   `json:"precipitation_10"`
		Precipitation30     float64   `json:"precipitation_30"`
		Precipitation60     float64   `json:"precipitation_60"`
		PressureMsl         float64   `json:"pressure_msl"`
		RelativeHumidity    float64   `json:"relative_humidity"`
		Visibility          int       `json:"visibility"`
		WindDirection10     int       `json:"wind_direction_10"`
		WindDirection30     int       `json:"wind_direction_30"`
		WindDirection60     int       `json:"wind_direction_60"`
		WindSpeed10         float64   `json:"wind_speed_10"`
		WindSpeed30         float64   `json:"wind_speed_30"`
		WindSpeed60         float64   `json:"wind_speed_60"`
		WindGustDirection10 int       `json:"wind_gust_direction_10"`
		WindGustDirection30 int       `json:"wind_gust_direction_30"`
		WindGustDirection60 int       `json:"wind_gust_direction_60"`
		WindGustSpeed10     float64   `json:"wind_gust_speed_10"`
		WindGustSpeed30     float64   `json:"wind_gust_speed_30"`
		WindGustSpeed60     float64   `json:"wind_gust_speed_60"`
		Sunshine30          float64   `json:"sunshine_30"`
		Sunshine60          float64   `json:"sunshine_60"`
		Temperature         float64   `json:"temperature"`
		FallbackSourceIds   struct {
		} `json:"fallback_source_ids"`
		Icon string `json:"icon"`
	} `json:"weather"`
	Sources []struct {
		ID              int       `json:"id"`
		DwdStationID    string    `json:"dwd_station_id"`
		ObservationType string    `json:"observation_type"`
		Lat             float64   `json:"lat"`
		Lon             float64   `json:"lon"`
		Height          float64   `json:"height"`
		StationName     string    `json:"station_name"`
		WmoStationID    string    `json:"wmo_station_id"`
		FirstRecord     time.Time `json:"first_record"`
		LastRecord      time.Time `json:"last_record"`
		Distance        float64   `json:"distance"`
	} `json:"sources"`
}

type WeatherUi struct {
	screen *ebiten.Image
}

var weatherCurrentData *BrightskyCurrentRes
var weatherPredictionData *BrightskyPredictionRes

func pollWeather() {
	for {
		err := fetchWeatherCurrent()
		if err != nil {
			fmt.Println("Error fetching current weather data:", err)
		}

		err = fetchWeatherPrediction()
		if err != nil {
			fmt.Println("Error fetching prediction weather data:", err)
		}

		time.Sleep(5 * time.Minute)
	}
}

func fetchWeatherCurrent() error {
	req, err := http.Get(
		"https://api.brightsky.dev/current_weather?lat=51.1857454&lon=6.90171",
	)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	var data BrightskyCurrentRes
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	weatherCurrentData = &data

	return nil
}

func fetchWeatherPrediction() error {
	req, err := http.Get(
		fmt.Sprintf(
			"https://api.brightsky.dev/weather?lat=51.1857454&lon=6.90171&date=%s",
			time.Now().Format("2006-01-02"),
		),
	)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	var data BrightskyPredictionRes
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	weatherPredictionData = &data

	return nil
}

func (ui *WeatherUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *WeatherUi) Bounds() (width, height int) {
	return WIDTH, fontHeight * 6
}

func (ui *WeatherUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	if weatherCurrentData == nil {
		return ui.screen
	}

	// Find closest weather data
	// currentWeather := weatherData.Weather[0]
	// for _, w := range weatherData.Weather {
	// 	if w.Timestamp.After(time.Now()) {
	// 		currentWeather = w
	// 		break
	// 	}
	// }

	// Draw weather text
	weatherS := fmt.Sprintf(
		"%.1f°c\n%s",
		weatherCurrentData.Weather.Temperature,
		weatherCurrentData.Weather.Condition,
	)
	text.Draw(ui.screen, weatherS, defaultFont, fontWidth*2, fontHeight, textColor)
	// Draw weather icon
	text.Draw(
		ui.screen,
		icon2Char(weatherCurrentData.Weather.Icon),
		weatherFont,
		fontWidth*6,
		fontHeight*4,
		textColor,
	)

	// Draw pollen
	pollenS := ""
	pollenKeys := []string{"g", "b", "h"}
	for _, key := range pollenKeys {
		v := pollenStrength[key]
		if v == "0" {
			continue
		}
		pollenS += fmt.Sprintf("%s%s\n", key, v)
	}
	text.Draw(ui.screen, pollenS, defaultFont, fontWidth*12, fontHeight*4, textColor)

	return ui.screen
}

func icon2Char(icon string) string {
	switch icon {

	case "cloudy":
		return ""
	case "partly-cloudy-day":
		return ""
	case "partly-cloudy-night":
		return ""
	case "clear-day":
		return ""
	case "clear-night":
		return ""
	case "rain":
		return ""
	case "snow":
		return ""
	case "sleet":
		return ""
	case "wind":
		return ""

	default:
		fmt.Println("Unknown icon", icon)
		return ""
	}
}
