package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rustyeddy/redeye"
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
	flag.IntVar(&config.VideoDevice, "device", 0, "Video capture device. default 0")
	flag.StringVar(&config.Image, "image", "", "Image name")
	flag.StringVar(&config.Video, "video", "", "Video Name")
}

func main() {
	flag.Parse()

	// list filters and exit if command list says so
	if config.ListFilters {
		listFilters()
		os.Exit(0)
	}

	// Determine the imgsrc
	imgsrc := startImgSrc(config)
	defer imgsrc.Close()

	// Set up the pipeline
	pipeline := redeye.NewPipeline(config.Pipeline)
	defer pipeline.Close()

	// Start the outputs windows and MJPEG server
	w := startWindows(config)
	defer w.Close()

	mjpeg := startMJPEG()
	defer mjpeg.Close()

	startServer()

	var outputs []chan *redeye.Frame
	outputs = append(outputs, mjpeg.Play())
	outputs = append(outputs, w.Play())

	frameQ := imgsrc.Play()
	for imgsrc.IsRunning() {
		f := <-frameQ
		for _, flt := range pipeline.Filters {
			f = flt.Filter(f)
		}
		for _, outQ := range outputs {
			outQ <- f
		}
	}
}

func startImgSrc(config *redeye.Configuration) (imgsrc redeye.ImgSrc) {
	var err error

	config.WaitTime = 10
	if config.Image != "" {
		imgsrc, err = redeye.GetImg(config.Image)
		config.WaitTime = 0
	} else if config.Video != "" {
		imgsrc, err = redeye.GetVideo(config.Video)
	} else {
		imgsrc, err = redeye.GetCam(config.VideoDevice)
	}
	if err != nil {
		log.Printf("Failed to open video device: %d - %+v", config.VideoDevice, err)
		os.Exit(1)
	}
	fmt.Println("returning waitTime", config.WaitTime)
	return imgsrc
}

func startWindows(config *redeye.Configuration) (w *redeye.Window) {
	w = redeye.NewWindow("Redeye")
	w.WaitTime = config.WaitTime
	return w
}

func startServer() *http.Server {
	server := &http.Server{
		Addr: config.HTTPAddr,
	}

	go server.ListenAndServe()
	return server
}

func startMJPEG() *redeye.MJPEG {
	// Create the MJPEG Stream, should this just be
	// a filter?
	mjpeg := redeye.NewMJPEG()
	http.Handle("/mjpeg", mjpeg)
	return mjpeg
}

func listFilters() {
	fmt.Println("Filters")
	names := redeye.Filters.List()
	for _, n := range names {
		flt, ok := redeye.Filters.Get(n)
		if !ok {
			fmt.Println("Bad filtername name: ", n)
			continue
		}
		fmt.Printf("%15s: %s\n", n, flt.Desc())
	}
}
