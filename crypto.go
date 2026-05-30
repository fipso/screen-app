package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const FIAT_SYMBOL = "USDT"

var pricesText string

// historyRetention is how far back we keep WS kline events. calcDelta only
// looks back 24h; the extra hour is slack so a delta lookup at the boundary
// still finds something.
const historyRetention = 25 * time.Hour

type CurrencyPair struct {
	mu      sync.Mutex
	symbol1 string
	symbol2 string
	price   float64
	history []binance.WsKlineEvent
}

type CryptoUi struct {
	screen *ebiten.Image
}

var (
	pairs   []*CurrencyPair
	symbols = []string{"BTC", "ETH", "SOL", "XRP", "APE", "RAY", "IOTA", "PEPE", "SHIB", "DOGE", "APT", "ADA", "BNB", "Link", "EUR"}
)

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
	// Snapshot prices once under each pair's lock so the sort sees a stable
	// view and the WS goroutine isn't racing the comparator.
	type snap struct {
		pair  *CurrencyPair
		price float64
	}
	snaps := make([]snap, 0, len(pairs))
	for _, pair := range pairs {
		pair.mu.Lock()
		snaps = append(snaps, snap{pair: pair, price: pair.price})
		pair.mu.Unlock()
	}
	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].price > snaps[j].price
	})
	out := make([]*CurrencyPair, len(snaps))
	for i, s := range snaps {
		out[i] = s.pair
	}
	return out
}

func watchCurrency(pair *CurrencyPair) {
	for {
		wsKlineHandler := func(event *binance.WsKlineEvent) {
			price, err := strconv.ParseFloat(event.Kline.Close, 64)
			if err != nil {
				fmt.Println(err)
			}
			pair.mu.Lock()
			pair.price = price
			pair.history = append(pair.history, *event)
			cutoffMs := time.Now().Add(-historyRetention).UnixMilli()
			drop := sort.Search(len(pair.history), func(i int) bool {
				return pair.history[i].Kline.EndTime >= cutoffMs
			})
			if drop > 0 {
				// Re-slice into a fresh backing array so the dropped prefix can be
				// GC'd; otherwise the underlying array keeps growing forever.
				kept := make([]binance.WsKlineEvent, len(pair.history)-drop)
				copy(kept, pair.history[drop:])
				pair.history = kept
			}
			pair.mu.Unlock()
		}
		doneC, _, err := binance.WsKlineServe(
			fmt.Sprintf("%s%s", pair.symbol1, pair.symbol2),
			"1m",
			wsKlineHandler,
			func(err error) {
				fmt.Println(err)
			},
		)
		if err != nil {
			fmt.Println("[BINANCE WS]", err)
			time.Sleep(time.Second * 5)
			continue
		}
		<-doneC
	}
}

func calcDelta(pair *CurrencyPair, span time.Duration) float64 {
	pair.mu.Lock()
	defer pair.mu.Unlock()

	if len(pair.history) == 0 {
		return 0
	}

	cutoffMs := time.Now().Add(-span).UnixMilli()
	idx := sort.Search(len(pair.history), func(i int) bool {
		return pair.history[i].Kline.EndTime >= cutoffMs
	})

	var s string
	if idx < len(pair.history) {
		s = pair.history[idx].Kline.Close
	} else {
		// No event within the span — fall back to the most recent one.
		s = pair.history[len(pair.history)-1].Kline.Close
	}
	f, err := strconv.ParseFloat(s, 64)
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
	return config.Width, sectionHeaderHeight + (fontHeight+linePadding)*len(symbols) + linePadding*4
}

func (ui *CryptoUi) Draw() *ebiten.Image {
	ui.screen.Fill(bgColor)

	contentY := drawSectionHeader(ui.screen, "markets · live", "24h", 0)
	contentWidth := config.Width - 2*paddingX

	prices := sortedCurrencyPairs()
	for i, currency := range prices {
		c := textColor
		delta := calcDelta(currency, time.Hour*24)
		dir := 0
		switch {
		case delta > 0:
			c = posColor
			dir = 1
		case delta < 0:
			c = negColor
			dir = -1
		}

		currency.mu.Lock()
		price := currency.price
		currency.mu.Unlock()

		value := fmt.Sprintf("%.2f", price)
		if price > 1000 {
			value = fmt.Sprintf("%.2fk", price/1000)
		}
		if price < 0.01 {
			value = fmt.Sprintf("%.2e", price)
		}

		pct := 0.0
		if price != 0 {
			pct = math.Abs(delta / price * 100)
		}

		y := contentY + fontHeight + (fontHeight+linePadding)*i
		prefix := fmt.Sprintf("%-5s %-8s", strings.ToLower(currency.symbol1), value)
		text.Draw(ui.screen, prefix, defaultFont, 0, y, textColor)

		deltaStr := fmt.Sprintf("%.1f%%", pct)
		b := text.BoundString(smallFont, deltaStr)
		deltaX := contentWidth - b.Dx()
		text.Draw(ui.screen, deltaStr, smallFont, deltaX, y, c)

		if dir != 0 {
			const triSize = 20
			triRight := float32(deltaX - 12)
			triLeft := triRight - triSize
			triCenter := (triLeft + triRight) / 2
			// Vertically align the triangle on the text x-height (~baseline - 14px).
			triMid := float32(y) - 14
			triHalf := float32(triSize) * 0.5
			if dir > 0 {
				drawFilledTriangle(ui.screen,
					triCenter, triMid-triHalf,
					triLeft, triMid+triHalf,
					triRight, triMid+triHalf,
					c)
			} else {
				drawFilledTriangle(ui.screen,
					triLeft, triMid-triHalf,
					triRight, triMid-triHalf,
					triCenter, triMid+triHalf,
					c)
			}
		}
	}

	return ui.screen
}

var (
	whitePixelImage = func() *ebiten.Image {
		img := ebiten.NewImage(3, 3)
		pix := make([]byte, 4*3*3)
		for i := range pix {
			pix[i] = 0xff
		}
		img.WritePixels(pix)
		return img
	}()
	whitePixelSub = whitePixelImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func drawFilledTriangle(dst *ebiten.Image, x0, y0, x1, y1, x2, y2 float32, clr color.Color) {
	var path vector.Path
	path.MoveTo(x0, y0)
	path.LineTo(x1, y1)
	path.LineTo(x2, y2)
	path.Close()
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)

	r, g, b, a := clr.RGBA()
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(r) / 0xffff
		vs[i].ColorG = float32(g) / 0xffff
		vs[i].ColorB = float32(b) / 0xffff
		vs[i].ColorA = float32(a) / 0xffff
	}
	op := &ebiten.DrawTrianglesOptions{}
	op.ColorScaleMode = ebiten.ColorScaleModePremultipliedAlpha
	op.AntiAlias = true
	dst.DrawTriangles(vs, is, whitePixelSub, op)
}
