package main

var config *Config

func main() {
	go pollBinance()
	go pollBusTimes()
	go pollPollen()
        go pollWeather()


        loadConfig()
	runGameUI()
}
