package main

import "github.com/fipso/screen-app/game"

func main() {
	go game.SubscribeBinance()
	go game.PollBusTimes()
	go game.PollPollen()
	go game.PollWeather()

	g := game.SetupGameUI()
	game.RunGame(g)
}
