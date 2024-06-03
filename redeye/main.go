package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rustyeddy/redeye"
	"github.com/rustyeddy/redeye/filters"

	"gocv.io/x/gocv"
)

var (
	config redeye.Configuration
)

func init() {
	flag.IntVar(&config.VideoDevice, "video-device", 0, "Video capture device. default 0")
	flag.StringVar(&config.CascadeFile, "cascade-file", "/usr/local/share/opencv4/haarcascades/haarcascade_frontalface_default.xml", "cascade file")
	flag.BoolVar(&config.ListFilters, "filters", true, "list available filters")
}

func main() {
	flag.Parse()

	if config.ListFilters {
		listFilters()
		os.Exit(0)
	}

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

func listFilters() {
	fmt.Println("Filters")
	names := filters.Filters.List()
	for _, n := range names {
		flt, ok := filters.Filters.Get(n)
		if !ok {
			fmt.Println("Bad filtername name: ", n)
			continue
		}
		fmt.Printf("%15s: %s\n", n, flt.Description())
	}
}
