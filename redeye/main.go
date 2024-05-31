package main

import (
	"flag"
	"log"
	"os"

	"gocv.io/x/gocv"
	"github.com/rustyeddy/redeye"
)

var (
	config redeye.Configuration
)

func init() {
	flag.IntVar(&config.VideoDevice, "video-device", 0, "Video capture device. default 0")
}

func main() {
	flag.Parse()

	cam, err := redeye.GetWebcam(config.VideoDevice)
	if err != nil {
		log.Printf("Failed to open video device: %d - %+v", config.VideoDevice, err)
		os.Exit(1)
	}
	window := gocv.NewWindow("Hello")

	imgQ := cam.Play()
	for {
		img := <-imgQ
		window.IMShow(*img)
		window.WaitKey(1)
	}
}



