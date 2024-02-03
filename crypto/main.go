package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/adshao/go-binance/v2"
)

const FIAT_SYMBOL = "USDT"

var pricesUpdated chan bool

type Currency struct {
	name  string
	price float64
}

var currencies = map[string]*Currency{
	"BTC":  {"Bitcoin", 0},
	"ETH":  {"Ethereum", 0},
	"SOL":  {"Solana", 0},
	"XMR":  {"Monero", 0},
	"XRP":  {"Ripple", 0},
	"APE":  {"Ape Coin", 0},
	"RNDR": {"Render", 0},
	"RAY":  {"Raydium", 0},
	"IOTA": {"Miota", 0},
}

func main() {
	pricesUpdated = make(chan bool)

	for symbol := range currencies {
		go watchCurrency(symbol)
	}

	setupTUI()
}

func watchCurrency(symbol string) {
	wsKlineHandler := func(event *binance.WsKlineEvent) {
		var err error
		currency := currencies[strings.Replace(symbol, FIAT_SYMBOL, "", 1)]
		currency.price, err = strconv.ParseFloat(event.Kline.Close, 64)
		if err != nil {
			fmt.Println(err)
		}

		pricesUpdated <- true
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
