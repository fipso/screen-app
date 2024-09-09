package game

import (
	"bytes"
	"image/jpeg"
	"log"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtpmjpeg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pion/rtp"
)

type RtspUi struct {
	screen *ebiten.Image

	currentImage *ebiten.Image
	scale        float64
	streamWidth  int
	streamHeight int
}

func (ui *RtspUi) Init() {
	ui.scale = 0.8
	ui.streamWidth = 1920
	ui.streamHeight = 1080
	width, height := ui.Bounds()
	ui.screen = ebiten.NewImage(width, height)

	c := gortsplib.Client{}

	// parse URL
	u, err := base.ParseURL("rtsp://192.168.178.78:8554/cam")
	if err != nil {
		log.Fatal(err)
	}

	// connect to the server
	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		log.Fatal(err)
	}

	// find available medias
	desc, _, err := c.Describe(u)
	if err != nil {
		log.Fatal(err)
	}

	// find the M-JPEG media and format
	var forma *format.MJPEG
	medi := desc.FindFormat(&forma)
	if medi == nil {
		log.Fatal("media not found")
	}

	// create decoder
	rtpDec, err := forma.CreateDecoder()
	if err != nil {
		log.Fatal(err)
	}

	// setup a single media
	_, err = c.Setup(desc.BaseURL, medi, 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	// called when a RTP packet arrives
	c.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
		// decode timestamp
		_, ok := c.PacketPTS(medi, pkt)
		if !ok {
			log.Printf("waiting for timestamp")
			return
		}

		// extract JPEG images from RTP packets
		enc, err := rtpDec.Decode(pkt)
		if err != nil {
			if err != rtpmjpeg.ErrNonStartingPacketAndNoPrevious && err != rtpmjpeg.ErrMorePacketsNeeded {
				log.Printf("ERR: %v", err)
			}
			return
		}

		// convert JPEG images into raw images
		newImg, err := jpeg.Decode(bytes.NewReader(enc))
		if err != nil {
			log.Fatal(err)
		}

		//log.Printf("decoded image with PTS %v and size %v", pts, newImg.Bounds().Max)

		ui.currentImage = ebiten.NewImageFromImage(newImg)
	})

	// start playing
	_, err = c.Play(nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (ui *RtspUi) Bounds() (width, height int) {
	return int(float64(ui.streamWidth) * ui.scale), int(float64(ui.streamHeight) * ui.scale)
}

func (ui *RtspUi) Draw() *ebiten.Image {
	if ui.currentImage == nil {
		return ui.screen
	}

	opts := &ebiten.DrawImageOptions{}

	opts.GeoM.Scale(ui.scale, ui.scale)

	ui.screen.DrawImage(ui.currentImage, opts)
	return ui.screen
}
