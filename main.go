package main

const SCALE = 2

func main() {
	go pollBinance()
	go pollBusTimes()

	runGameUI()
}
