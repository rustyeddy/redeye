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
	config *redeye.Configuration
)

func init() {
	config = redeye.GetConfig()
	flag.StringVar(&config.CascadeFile, "cascade-file", "/usr/local/share/opencv4/haarcascades/haarcascade_frontalface_default.xml", "cascade file")
	flag.BoolVar(&config.ListFilters, "filters", false, "list available filters")
	flag.StringVar(&config.Pipeline, "pipeline", "", "list of fliters separated by colons")
	flag.IntVar(&config.VideoDevice, "video-device", 0, "Video capture device. default 0")
	flag.StringVar(&config.Imgname, "img", "", "Image name")
}

func main() {
	flag.Parse()

	// list filters and exit if command list says so
	if config.ListFilters {
		listFilters()
		os.Exit(0)
	}

	var imgsrc redeye.ImgSrc
	var err error

	if config.Imgname != "" {
		imgsrc, err = redeye.GetImg(config.Imgname)
	} else {
		imgsrc, err = redeye.GetCam(config.VideoDevice)
	}
	if err != nil {
		log.Printf("Failed to open video device: %d - %+v", config.VideoDevice, err)
		os.Exit(1)
	}
	defer imgsrc.Close()

	pipeline := filters.NewPipeline(config.Pipeline)
	window := gocv.NewWindow("Redeye")
	window.ResizeWindow(640, 480)
	defer window.Close()

	frameQ := imgsrc.Play()
	for imgsrc.IsRunning() {
		f := <-frameQ
		for _, flt := range pipeline.Filters {
			f = flt.Filter(f)
		}

		fmt.Printf("mat: %d %d\n", f.Mat.Rows, f.Mat.Cols)
		window.IMShow(*f.Mat)
		window.WaitKey(-1)
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
		fmt.Printf("%15s: %s\n", n, flt.Desc())
	}
}
