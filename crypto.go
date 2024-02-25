package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
)

const FIAT_SYMBOL = "USDT"

var pricesText string

type Currency struct {
	name    string
	price   float64
	history []binance.WsKlineEvent
}

var currencies = map[string]*Currency{
	"BTC":  {"Bitcoin", 0, make([]binance.WsKlineEvent, 0)},
	"ETH":  {"Ethereum", 0, make([]binance.WsKlineEvent, 0)},
	"SOL":  {"Solana", 0, make([]binance.WsKlineEvent, 0)},
	"XMR":  {"Monero", 0, make([]binance.WsKlineEvent, 0)},
	"XRP":  {"Ripple", 0, make([]binance.WsKlineEvent, 0)},
	"APE":  {"Ape Coin", 0, make([]binance.WsKlineEvent, 0)},
	"RNDR": {"Render", 0, make([]binance.WsKlineEvent, 0)},
	"RAY":  {"Raydium", 0, make([]binance.WsKlineEvent, 0)},
	"IOTA": {"Miota", 0, make([]binance.WsKlineEvent, 0)},
}

func pollBinance() {
	for symbol := range currencies {
		go watchCurrency(symbol)
	}
}

func sortedPrices() []*Currency {
	// Sort cyrrencies by price
	var sortedCurrencies []*Currency
	for _, currency := range currencies {
		sortedCurrencies = append(sortedCurrencies, currency)
	}
	for i := 0; i < len(sortedCurrencies); i++ {
		for j := i + 1; j < len(sortedCurrencies); j++ {
			if sortedCurrencies[i].price < sortedCurrencies[j].price {
				sortedCurrencies[i], sortedCurrencies[j] = sortedCurrencies[j], sortedCurrencies[i]
			}
		}
	}

	return sortedCurrencies
}

func watchCurrency(symbol string) {
	wsKlineHandler := func(event *binance.WsKlineEvent) {
		var err error
		currency := currencies[strings.Replace(symbol, FIAT_SYMBOL, "", 1)]
		currency.price, err = strconv.ParseFloat(event.Kline.Close, 64)
		currency.history = append(currency.history, *event)
		if err != nil {
			fmt.Println(err)
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, _, err := binance.WsKlineServe(
		fmt.Sprintf("%s%s", symbol, FIAT_SYMBOL),
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

func oneHourDelta(currency *Currency) float64 {
	if len(currency.history) == 0 {
		return 0
	}

	for _, event := range currency.history {
		if time.Now().Sub(time.Unix(event.Kline.EndTime/1000, 0)) < time.Hour {
			s := event.Kline.Close
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				log.Fatal(err)
			}

			return currency.price - f
		}
	}

	// If no event in the last hour, return the delta between the last event and the current price
	last := currency.history[len(currency.history)-1].Kline.Close
	f, err := strconv.ParseFloat(last, 64)
	if err != nil {
		log.Fatal(err)
	}
	return currency.price - f
}
