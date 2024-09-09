package screenappmobile

import (
	"github.com/fipso/screen-app/game"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

func init() {
    mobile.SetGame(game.SetupGameUI())
}

// Dummy forces gomobile to compile this package.
func Dummy() {}
