package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/adshao/go-binance/v2"
)

const FIAT_SYMBOL = "USDT"

var currencies = []string{"BTC", "ETH", "SOL", "XMR", "XRP", "APE", "RNDR", "RAY", "IOTA"}
var prices = []float64{0, 0, 0, 0, 0, 0, 0, 0, 0}

func main() {

	a := app.New()
	w := a.NewWindow("Screen App")
	//w.SetFullScreen(true)

	list := widget.NewList(
		func() int {
			return len(currencies)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fmt.Sprintf("%s %2.f $", currencies[i], prices[i]))
		})

	w.SetContent(list)

	for i := range currencies {
		go watchCurrency(i)
	}

	w.ShowAndRun()
}

func watchCurrency(index int) {
	symbol := currencies[index]

	wsKlineHandler := func(event *binance.WsKlineEvent) {
		fmt.Println(event.Symbol, event.Kline.Close)

		var err error
		prices[index], err = strconv.ParseFloat(event.Kline.Close, 64)
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
