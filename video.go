package main

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
)

func readVideo(fileName string, writer io.WriteCloser) error {
	err := ffmpeg.Input(fileName).Output("pipe:", ffmpeg.KwArgs{
		"format":  "rawvideo",
		"pix_fmt": "rgb24",
	}).WithOutput(writer).ErrorToStdOut().Run()

	if err != nil {
		return err
	}

	return nil
}
