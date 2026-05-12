package main

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

var (
	audioCtx     *audio.Context
	audioCtxOnce sync.Once
)

func getAudioContext() *audio.Context {
	audioCtxOnce.Do(func() {
		audioCtx = audio.NewContext(44100)
	})
	return audioCtx
}
