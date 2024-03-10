package main

func main() {
	go pollBinance()
	go pollBusTimes()
	go pollPollen()

	runGameUI()
}
