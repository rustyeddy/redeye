package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/rustyeddy/redeye"
	"github.com/rustyeddy/redeye/ocv"
)

type Config struct {
	Addr   string
	Filter string
	Device interface{}
}

var config Config

func init() {
	flag.StringVar(&config.Addr, "addr", ":1234", "Listent to address")
	flag.StringVar(&config.Filter, "filter", "", "Filter to apply")
}

// go:embed index.html
// var content embed.FS

func main() {
	flag.Parse()

	srv := redeye.NewWebServer(config.Addr)
	srv.Handle("/", http.FileServer(http.Dir("./html")))

	if len(os.Args) < 1 {
		log.Println("No video capture devices specified")
		return
	}

	vidsrcs := getVideoSrcs(os.Args[1:])
	for i, vsrc := range vidsrcs {
		mjpg := redeye.NewMJPEGPlayer(i)
		srv.ServeMux.Handle(mjpg.GetURL(), mjpg)

		imgQ := vsrc.Play()
		mjpg.Play(imgQ)

	}
	srv.Listen()
}

func getVideoSrcs(args []string) []*VideoSource {

	devnum := 0
	var capdevs []*ocv.CaptureDevice
	for _, capstr := range os.Args[1:] {

		// Open up the video capture device
		cap := ocv.GetVideoSource(capstr)
		if cap == nil {
			log.Println("Failed to get capture device", capstr)
			os.Exit(1)
		}
		capdevs = append(capdevs, cap)
	}
	return capdevs

}
