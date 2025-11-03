package main

import (
	"image/color"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
)

type UiElement interface {
	Init()
	Bounds() (width, height int)
	Draw() *ebiten.Image
}

// Ebiten units
// const config.Width = 360 / 2
// const config.Height = 640 / 2
var (
	fontHeight = 72
	fontWidth  = 60
)

var (
	defaultFont font.Face = basicfont.Face7x13
	weatherFont font.Face = basicfont.Face7x13
	clockFont   font.Face = basicfont.Face7x13
	tinyFont    font.Face = basicfont.Face7x13
	smallFont   font.Face = basicfont.Face7x13
	faFont      font.Face = basicfont.Face7x13
)

var (
	textColor = color.RGBA{255, 255, 255, 255}
	bgColor   = color.RGBA{0, 0, 0, 255}
)

var (
	linePadding = 5
	paddingX    = 40
)

type Game struct {
	stackLayout  []UiElement
	currentModal UiElement
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(bgColor)

	// Draw content
	drawStackLayout(screen, g.stackLayout)

	// Draw modal if any
	if g.currentModal != nil {
		modalOverlay := g.currentModal.Draw()
		// Draw ontop of existing screen (with alpha transparency)
		screen.DrawImage(modalOverlay, &ebiten.DrawImageOptions{})
	}
}

func drawStackLayout(target *ebiten.Image, elements []UiElement) {
	// Draw UI elements
	pos := ebiten.GeoM{}
	pos.Translate(float64(paddingX), 0)
	for _, ui := range elements {
		img := ui.Draw()
		target.DrawImage(img, &ebiten.DrawImageOptions{
			GeoM: pos,
		})

		// Move to the next position
		pos.Translate(0, float64(img.Bounds().Dy()+(linePadding)))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// s := ebiten.DeviceScaleFactor()
	// return int(float64(outsideWidth) * s), int(float64(outsideHeight) * s)
	return config.Width, config.Height
	// return 1080, 1920
}

func parseUiElement(configElem LayoutElement) *UiElement {
	var element UiElement
	switch configElem.Type {
	case LayoutElementSwitch:
		var children []UiElement
		for _, configChild := range configElem.Children {
			child := parseUiElement(configChild)
			children = append(children, *child)
		}
		element = &SwitchLayout{
			interval: configElem.SwitchInterval,
			children: children,
		}
	case LayoutElementGrow:
		element = &GrowUi{}
	case LayoutElementBus:
		element = &BusUi{}
	case LayoutElementWeather:
		element = &WeatherUi{}
	case LayoutElementKnife:
		element = &KnifeAttackUi{}
	case LayoutElementClock:
		element = &ClockUi{}
	case LayoutElementCrypto:
		element = &CryptoUi{}
	case LayoutElementEnergy:
		element = &EnergyUi{}
	default:
		log.Fatalf("CONFIG | Unknown layout element type: %s", configElem.Type)
	}

	return &element
}

func runGameUI() {
	game := &Game{}

	// Build UI Layout from config
	for _, layoutElement := range config.Layout {
		element := parseUiElement(layoutElement)
		game.stackLayout = append(game.stackLayout, *element)
	}

	// Load UI elements
	/*
		game.stackLayout = append(game.stackLayout, &ClockUi{})

		// SwitchLayouts
		switchLayout := &SwitchLayout{
			interval: 2000, // 2k frames
			children: []UiElement{
				&CryptoUi{},
				&KnifeAttackUi{},
				//&GrowUi{},
			},
		}
		if config.Grow_mqtt.Enabled {
			switchLayout.children = append(switchLayout.children, &GrowUi{})
		}
		game.stackLayout = append(game.stackLayout, switchLayout)

		//game.stackLayout = append(game.stackLayout, &GrowUi{})
		//game.stackLayout = append(game.stackLayout, &GrowUi{})

		// game.stackLayout = append(game.stackLayout, &BusUi{})
		// game.stackLayout = append(game.stackLayout, &PollenUi{})
		switchLayout2 := &SwitchLayout{
			interval: 600, // 600 frames
			children: []UiElement{
				&BusUi{},
				&WeatherUi{},
			},
		}
		game.stackLayout = append(game.stackLayout, switchLayout2)*/
	// game.stackLayout = append(game.stackLayout, &PollenUi{})

	for _, ui := range game.stackLayout {
		ui.Init()
	}

	// DEBUG:!!!
	// Spawn test modal

	game.currentModal = &ModalUi{
		stackLayout: []UiElement{
			&AlertUi{
				msg: "  faggot on the\n     doooooor",
			},
		},
	}
	game.currentModal.Init()

	// Dark/Light mode
	go func() {
		for {
			if time.Now().Hour() > 17 || time.Now().Hour() < 8 {
				textColor = color.RGBA{255, 255, 255, 255}
				bgColor = color.RGBA{0, 0, 0, 255}
			} else {
				textColor = color.RGBA{0, 0, 0, 255}
				bgColor = color.RGBA{245, 245, 245, 255}
			}

			time.Sleep(time.Second)
		}
	}()

	fontHeight = config.Default_Font_Size
	// Load font
	defaultFont = loadFont("assets/fonts/MajorMonoDisplay-Regular.ttf", float64(config.Default_Font_Size))
	weatherFont = loadFont("assets/fonts/weathericons-regular-webfont.ttf", 260)
	clockFont = loadFont("assets/fonts/technology.bold.ttf", 100)
	tinyFont = loadFont("assets/fonts/OpenSans-Regular.ttf", 32)
	smallFont = loadFont("assets/fonts/OpenSans-Regular.ttf", 48)
	faFont = loadFont("assets/fonts/fa400.otf", 48*2)

	ebiten.SetWindowSize(config.Width, config.Height)
	ebiten.SetWindowTitle("screen-app ")
	ebiten.SetFullscreen(config.Fullscreen)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func loadFont(path string, size float64) font.Face {
	fontData, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	tt, err := opentype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}

	f, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})
	if err != nil {
		log.Fatal(err)
	}

	return f
}
