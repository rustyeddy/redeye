package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rustyeddy/redeye"
	"github.com/rustyeddy/redeye/filters"
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
	flag.StringVar(&config.Video, "video", "", "Video file name")
	flag.StringVar(&config.RTSPUrl, "rtsp", "", "RTSP stream URL (e.g. rtsp://camera.local/stream)")
}

func main() {
	if err := config.LoadDefault(); err != nil {
		log.Fatalf("failed to load default config: %+v", err)
	}

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
	pipeline := filters.NewPipeline(config.Pipeline)
	defer pipeline.Close()

	// Start the outputs windows and MJPEG server
	w := startWindows(config)
	defer w.Close()

	mjpeg := startMJPEG()
	defer mjpeg.Close()

	startServer()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		<-sigCh
		log.Println("shutting down ...")
		close(done)
	}()

	var outputs []chan *redeye.Frame
	outputs = append(outputs, mjpeg.Play())
	outputs = append(outputs, w.Play())

	frameQ := imgsrc.Play()
	for imgsrc.IsRunning() {
		select {
		case <-done:
			return
		case f, ok := <-frameQ:
			if !ok {
				return
			}
			for _, flt := range pipeline.Filters {
				f = flt.Filter(f)
			}
			var wg sync.WaitGroup
			for _, outQ := range outputs {
				wg.Add(1)
				go func(ch chan *redeye.Frame) {
					defer wg.Done()
					ch <- f
				}(outQ)
			}
			wg.Wait()
		}
	}

	// For a static image, hold the window open until the user presses a key
	// or the process is interrupted. Only the Window goroutine calls WaitKey;
	// main waits on the resulting signal channel to avoid concurrent highgui calls.
	if config.Image != "" {
		select {
		case <-done:
		case <-w.KeyPressed():
		}
	}
}

func startImgSrc(config *redeye.Configuration) (imgsrc redeye.ImgSrc) {
	var err error

	config.WaitTime = 10
	if config.Image != "" {
		imgsrc, err = redeye.GetImg(config.Image)
		config.WaitTime = 0
	} else if config.RTSPUrl != "" {
		imgsrc, err = redeye.GetRTSP(config.RTSPUrl)
	} else if config.Video != "" {
		imgsrc, err = redeye.GetVideo(config.Video)
	} else {
		imgsrc, err = redeye.GetCam(config.VideoDevice)
	}
	if err != nil {
		log.Printf("Failed to open video device: %d - %+v", config.VideoDevice, err)
		os.Exit(1)
	}
	return imgsrc
}

func startWindows(config *redeye.Configuration) (w *redeye.Window) {
	w = redeye.NewWindow("Redeye")
	w.WaitTime = config.WaitTime
	return w
}

func startServer() *http.Server {
	redeye.RegisterAPIRoutes(http.DefaultServeMux)

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
