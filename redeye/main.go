package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
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
	flag.StringVar(&config.HTTPAddr, "addr", "0.0.0.0:8080", "Default http addr 8080")
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

	// Set up the image source
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

	// Set up the pipeline
	pipeline := filters.NewPipeline(config.Pipeline)
	window := gocv.NewWindow("Redeye")
	window.ResizeWindow(640, 480)
	defer window.Close()

	// Create the MJPEG Stream, should this just be
	// a filter?
	mjpeg := redeye.NewMJPEG()
	http.Handle("/mjpeg", mjpeg)
	server := &http.Server{
		Addr: config.HTTPAddr,
	}

	go server.ListenAndServe()
	frameQ := imgsrc.Play()
	mjpegQ := mjpeg.Play()
	for imgsrc.IsRunning() {
		f := <-frameQ
		for _, flt := range pipeline.Filters {
			f = flt.Filter(f)
		}
		mjpegQ <- f
		window.IMShow(*f.Mat)
		window.WaitKey(10)
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
