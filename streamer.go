package redeye


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
import (
	"fmt"
	"log"
	"time"

	"gocv.io/x/gocv"
)

type VideoStreamer interface{
	Stream(deviceID interface{}, vidQ chan []byte)
}

type GOCVStreamer struct {
}

func (st *GOCVStreamer) Stream(deviceID interface{}, vidQ chan []byte) {

	go func() {
		// open webcam
		webcam, err := gocv.OpenVideoCapture(deviceID)
		if err != nil {
			fmt.Printf("Error opening capture device: %v\n", deviceID)
			return 
		}
		defer webcam.Close()
		img := gocv.NewMat()
		defer img.Close()

		for {
			if ok := webcam.Read(&img); !ok {
				log.Printf("Bad read:\n")
				time.Sleep(1 * time.Second)
				continue
			}
			if img.Empty() {
				log.Println("Empty image")
				continue
			}
			jpg, _ := gocv.IMEncode(".jpg", img)
			vidQ <- jpg.GetBytes()
			time.Sleep(5 * time.Millisecond)
			jpg.Close()
		}
	} ()
}
