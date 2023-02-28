package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rustyeddy/redeye"
	"github.com/rustyeddy/redeye/ocv"
)

type Config struct {
	Device interface{}
}

// go:embed index.html
// var content embed.FS

func main() {

	if len(os.Args) < 2 {
		log.Println("No video capture devices specified")
		return
	}

	// Move HTML to web
	host := ":1234"
	srv := redeye.NewWebServer(host)
	srv.Handle("/", http.FileServer(http.Dir("./html")))

	devnum := 0
	var capdevs []*ocv.CaptureDevice
	for _, capstr := range os.Args[1:] {

		// Open up the video capture device
		cap := ocv.GetCaptureDevice(capstr)
		if cap == nil {
			log.Println("Failed to get capture device", capstr)
			os.Exit(1)
		}
		capdevs = append(capdevs, cap)
		cap.Filter = ocv.NullFilter{}

		// Open the MJPEG player and register with the http server
		mjpg := redeye.NewMJPEGPlayer()
		url := fmt.Sprintf("/mjpeg/%d", devnum)
		srv.ServeMux.Handle(url, mjpg.Stream)
		devnum++
		log.Printf("Capture device: %v\n", cap.DeviceID)

		// create the channel to pump video from the capture device
		// to the MJPEG player
		vidQ := mjpg.Play()
		cap.Stream(vidQ)

		log.Println("Streaming video at ", host, url)
	}

	srv.Listen()
}
