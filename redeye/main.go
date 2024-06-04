package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
}

func main() {
	flag.Parse()

	// list filters and exit if command list says so
	if config.ListFilters {
		listFilters()
		os.Exit(0)
	}

	// Setup the pipeline for filtering
	if config.Pipeline != "" {
		setupPipeline(config.Pipeline)
	}

	// Open web cam for streaming video
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
	imgQ := cam.Play()
	outQ := filters.Pipes.Start(imgQ)

	for redeye.Running {
		fmt.Printf("Reading from outQ: %+v\n", outQ)
		img, redeye.Running = <-outQ
		window.IMShow(*img)
		window.WaitKey(1)
	}
}

func setupPipeline(pipestr string) {
	flts := strings.Split(pipestr, ":")
	for _, flt := range flts {
		fmt.Printf("filter: %s\n", flt)
		filters.Pipes.AddFilter(flt)
	}

	fmt.Printf("PIPES: %v\n", filters.Pipes)
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
