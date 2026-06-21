package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rustyeddy/redeye"
	_ "github.com/rustyeddy/redeye/filters/colors"
	_ "github.com/rustyeddy/redeye/filters/facedetect"
	_ "github.com/rustyeddy/redeye/filters/resize"
	_ "github.com/rustyeddy/redeye/filters/textzoom"
)

var (
	config *redeye.Configuration
)

func init() {
	config = redeye.GetConfig()
	flag.StringVar(&config.CascadeFile, "cascade-file", "/usr/local/share/opencv4/haarcascades/haarcascade_frontalface_default.xml", "cascade file")
	flag.StringVar(&config.HTTPAddr, "addr", "0.0.0.0:9382", "Default http addr 8080")
	flag.BoolVar(&config.ListFilters, "filters", false, "list available filters")
	flag.StringVar(&config.PluginDir, "plugins", "plugins", "directory of .so filter plugins to load at startup")
	flag.StringVar(&config.Pipeline, "pipeline", "", "list of fliters separated by colons")
	flag.StringVar(&config.VideoDevice, "device", "0", "Camera device: index (0,1,…), name (jetson,nano,rpi,linux,mac), or path (/dev/video0)")
	flag.StringVar(&config.VideoDevice, "camera", "0", "Camera to use (alias for --device)")
	flag.BoolVar(&config.ListCameras, "list-cameras", false, "print available cameras and exit")
	flag.StringVar(&config.Image, "image", "", "Image name")
	flag.StringVar(&config.Video, "video", "", "Video file name")
	flag.StringVar(&config.RTSPUrl, "rtsp", "", "RTSP stream URL (e.g. rtsp://camera.local/stream)")
	flag.StringVar(&config.MQTTBroker, "broker", "", "MQTT broker URL (e.g. tcp://localhost:1883); empty disables MQTT")
	flag.StringVar(&config.MQTTTopicPrefix, "topic-prefix", "/redeye", "MQTT topic namespace prefix")
	flag.StringVar(&config.LogFile, "log", "redeye.log", "log destination: stderr, stdout, or a file path")
	flag.StringVar(&config.LogLevel, "log-level", "info", "log level: debug, info, warn, error")
}

func main() {
	if err := config.LoadDefault(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load default config: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()
	logCloser, err := redeye.InitLogger(config.LogFile, config.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %v\n", err)
		os.Exit(1)
	}
	if logCloser != nil {
		defer logCloser.Close()
	}

	if config.ListCameras {
		for _, cam := range redeye.ListCameras() {
			if cam.Name != "" {
				fmt.Printf("%-14s  %s\n", cam.Device, cam.Name)
			} else {
				fmt.Println(cam.Device)
			}
		}
		os.Exit(0)
	}

	// Load any .so filter plugins from the plugin directory before building the pipeline.
	if n, err := redeye.LoadPlugins(config.PluginDir); err != nil {
		slog.Warn("plugin dir load error", "dir", config.PluginDir, "err", err)
	} else if n > 0 {
		slog.Info("plugins loaded", "count", n, "dir", config.PluginDir)
	}

	if config.ListFilters {
		listFilters()
		os.Exit(0)
	}

	// Determine the imgsrc
	imgsrc := startImgSrc(config)
	defer imgsrc.Close()

	// Register the source as the snap target if it supports snapping.
	if s, ok := imgsrc.(redeye.Snapper); ok {
		redeye.SetSnapper(s)
	}

	// Start MQTT messenger if a broker is configured.
	if config.MQTTBroker != "" {
		if msgr := startMessenger(config); msgr != nil {
			defer msgr.Close()
		}
	}

	// Set up the pipeline and register it as the active pipeline so it can
	// be swapped at runtime via REST or MQTT.
	pipeline, err := redeye.NewPipeline(config.Pipeline)
	if err != nil {
		slog.Error("invalid pipeline", "err", err)
		os.Exit(1)
	}
	redeye.SetPipeline(pipeline)

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
		slog.Info("shutting down")
		close(done)
	}()

	var outputs []chan *redeye.Frame
	outputs = append(outputs, mjpeg.Play())
	outputs = append(outputs, w.Play())

	runFrames(imgsrc, outputs, done)

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

// runFrames feeds frames from imgsrc through the active pipeline to all outputs.
// It listens for camera hot-swap requests on redeye.CameraSwitch(): on receipt
// it closes the current source, opens the new one, and resumes without restart.
func runFrames(imgsrc redeye.ImgSrc, outputs []chan *redeye.Frame, done <-chan struct{}) {
	for {
		frameQ := imgsrc.Play()
		newDevice := ""
	stream:
		for {
			select {
			case <-done:
				return
			case device := <-redeye.CameraSwitch():
				newDevice = device
				break stream
			case f, ok := <-frameQ:
				if !ok {
					return
				}
				for _, flt := range redeye.GetPipeline().Filters {
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

		// Close current camera first, then open the new one.
		imgsrc.Close()
		newCam, err := redeye.GetCam(newDevice)
		if err != nil {
			slog.Error("camera switch failed", "device", newDevice, "err", err)
			return
		}
		imgsrc = newCam
		redeye.Config.VideoDevice = newDevice
		if s, ok := imgsrc.(redeye.Snapper); ok {
			redeye.SetSnapper(s)
		}
		slog.Info("camera switched", "device", newDevice)
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
		if err != nil {
			slog.Warn("no camera available, starting without video", "device", config.VideoDevice, "err", err)
			return redeye.NewNullSrc()
		}
	}
	if err != nil {
		slog.Error("failed to open source", "err", err)
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

	slog.Info("http listening", "addr", config.HTTPAddr)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "err", err)
		}
	}()
	return server
}

func startMJPEG() *redeye.MJPEG {
	// Create the MJPEG Stream, should this just be
	// a filter?
	mjpeg := redeye.NewMJPEG()
	http.Handle("/mjpeg", mjpeg)
	return mjpeg
}

func startMessenger(config *redeye.Configuration) *redeye.Messenger {
	m := redeye.NewMessenger(config.MQTTBroker, config.MQTTTopicPrefix)
	if err := m.Connect(); err != nil {
		slog.Info("MQTT disabled", "err", err)
		m.Close()
		return nil
	}
	if err := m.Announce(); err != nil {
		slog.Warn("MQTT announce failed", "err", err)
	}
	if err := m.SubscribeCommands(); err != nil {
		slog.Warn("MQTT subscribe failed", "err", err)
	}
	return m
}

func listFilters() {
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
