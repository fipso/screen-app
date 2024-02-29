package main

const SCALE = 2

func main() {
	go pollBinance()
	go pollBusTimes()
	go pollPollen()

	runGameUI()
}
