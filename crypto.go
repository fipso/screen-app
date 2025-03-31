package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const FIAT_SYMBOL = "USDT"

var pricesText string

type CurrencyPair struct {
	symbol1 string
	symbol2 string
	price   float64
	history []binance.WsKlineEvent
}

type CryptoUi struct {
	screen *ebiten.Image
}

var pairs []*CurrencyPair
var symbols = []string{"BTC", "ETH", "SOL", "XRP", "APE", "RAY", "IOTA", "PEPE", "SHIB", "DOGE", "APT", "ADA", "BNB", "Link", "EUR"}

func pollBinance() {
	// default usdt pairs

	for _, symbol := range symbols {
		pair := &CurrencyPair{
			symbol1: symbol,
			symbol2: "USDT",
			price:   0,
			history: make([]binance.WsKlineEvent, 0),
		}
		pairs = append(pairs, pair)

		go watchCurrency(pair)
	}
}

func sortedCurrencyPairs() []*CurrencyPair {
	// Sort cyrrencies by price
	var sortedPairs []*CurrencyPair
	for _, pair := range pairs {
		sortedPairs = append(sortedPairs, pair)
	}
	for i := 0; i < len(sortedPairs); i++ {
		for j := i + 1; j < len(sortedPairs); j++ {
			if sortedPairs[i].price < sortedPairs[j].price {
				sortedPairs[i], sortedPairs[j] = sortedPairs[j], sortedPairs[i]
			}
		}
	}

	return sortedPairs
}

func watchCurrency(pair *CurrencyPair) {
	wsKlineHandler := func(event *binance.WsKlineEvent) {
		var err error
		pair.price, err = strconv.ParseFloat(event.Kline.Close, 64)
		pair.history = append(pair.history, *event)
		if err != nil {
			fmt.Println(err)
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, _, err := binance.WsKlineServe(
		fmt.Sprintf("%s%s", pair.symbol1, pair.symbol2),
		"1m",
		wsKlineHandler,
		errHandler,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	<-doneC
}

func calcDelta(pair *CurrencyPair, span time.Duration) float64 {
	if len(pair.history) == 0 {
		return 0
	}

	for _, event := range pair.history {
		if time.Now().Sub(time.Unix(event.Kline.EndTime/1000, 0)) < span {
			s := event.Kline.Close
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				log.Fatal(err)
			}

			return pair.price - f
		}
	}

	// If no event in the last hour, return the delta between the last event and the current price
	last := pair.history[len(pair.history)-1].Kline.Close
	f, err := strconv.ParseFloat(last, 64)
	if err != nil {
		log.Fatal(err)
	}
	return pair.price - f
}

func (ui *CryptoUi) Init() {
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)
}

func (ui *CryptoUi) Bounds() (width, height int) {
	return WIDTH, (fontHeight+linePadding)*len(symbols) + linePadding*10
}

func (ui *CryptoUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	prices := sortedCurrencyPairs()
	for i, currency := range prices {
		c := textColor
		delta := calcDelta(currency, time.Hour*24)
		if delta > 0 {
			c = color.RGBA{20, 200, 20, 255}
		} else if delta < 0 {
			c = color.RGBA{255, 0, 0, 255}
		}

		value := fmt.Sprintf("%.2f", currency.price)
		if currency.price > 1000 {
			value = fmt.Sprintf("%.2fk", currency.price/1000)
		}
		if currency.price < 0.01 {
			value = fmt.Sprintf("%.2e", currency.price)
		}

		line := fmt.Sprintf("%-5s %-8s %.1f%%", strings.ToLower(currency.symbol1), value, math.Abs(delta/currency.price*100))
		text.Draw(ui.screen, line, defaultFont, 0, (fontHeight+linePadding)*(i+1), c)
	}

	return ui.screen
}
