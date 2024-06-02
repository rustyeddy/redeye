package main

import (
	"flag"
	"log"
	"os"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

var (
	config redeye.Configuration
)

func main() {
	flag.Parse()

	cam, err := redeye.GetWebcam(config.VideoDevice)
	if err != nil {
		log.Printf("Failed to open video device: %d - %+v", config.VideoDevice, err)
		os.Exit(1)
	}
	defer cam.Close()

	window := gocv.NewWindow("Redeye")
	window.ResizeWindow(640, 480)
	defer window.Close()

	var img *gocv.Mat
	playing := true

	imgQ := cam.Play()
	for playing {
		img, playing = <-imgQ
		window.IMShow(*img)
		window.WaitKey(1)
	}
}
