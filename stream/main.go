// What it does:
//
// This example opens a video capture device, then streams MJPEG from it.
// Once running point your browser to the hostname/port you passed in the
// command line (for example http://localhost:8080) and you should see
// the live video stream.
//
// How to run:
//
// mjpeg-streamer [camera ID] [host:port]
//
//		go get -u github.com/hybridgroup/mjpeg
// 		go run ./cmd/mjpeg-streamer/main.go 1 0.0.0.0:8080
//

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

var (
	deviceID int
	err      error
	webcam   *gocv.VideoCapture
	stream   *mjpeg.Stream
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("How to run:\n\tmjpeg-streamer [camera ID] [host:port]")
		return
	}

	// parse args
	deviceID := "/dev/video0"
	host := ":1234"

	// open webcam
	webcam, err = gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()
	webcam.Set(gocv.VideoCaptureFrameWidth, 1280)
	webcam.Set(gocv.VideoCaptureFrameHeight, 720)

	// create the mjpeg stream
	stream = mjpeg.NewStream()

	img := gocv.NewMat()
	defer img.Close()

	fmt.Println("Capturing. Point your browser to " + host)

	// start http server
	http.Handle("/mjpeg", stream)

	imgQ := updateJPEG()

	go func() {
		for {
			if ok := webcam.Read(&img); !ok {
				fmt.Printf("Bad read: %v\n", deviceID)
				time.Sleep(1 * time.Second)
				continue
			}

			if img.Empty() {
				log.Println("Empty image")
				continue
			}

			imgQ <- &img
			time.Sleep(5 * time.Millisecond)
		}
	}()

	log.Fatal(http.ListenAndServe(host, nil))
}

func updateJPEG() chan *gocv.Mat {
	imgQ := make(chan *gocv.Mat)

	go func() {
		for {
			select {
			case img := <-imgQ:
				buf, _ := gocv.IMEncode(".jpg", *img)
				stream.UpdateJPEG(buf.GetBytes())
				buf.Close()
			}
		}
	}()

	return imgQ
}
