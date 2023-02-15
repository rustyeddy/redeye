package streamer


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
	"net/http"
	_ "os"
	"time"
	"strconv"

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

func jetsonCamstr(width int, height int, frame int, flip int) string {
	w := strconv.Itoa(width)
	h := strconv.Itoa(height)
	f := strconv.Itoa(frame)
	fl := strconv.Itoa(flip)

	str := "nvarguscamerasrc ! video/x-raw(memory:NVMM), width=(int)" + string(w) + ", height=(int)" +
		string(h) + ", framerate=(fraction)" + string(f) + "/1 ! nvvidconv flip-method=" +
		string(fl) + " ! video/x-raw, width=(int)" + string(w) +
		", height=(int)" + string(h) +
		", format=(string)BGRx ! videoconvert ! video/x-raw, format=(string)BGR ! appsink";
	return str
}



func streamer() {
	// parse args
	deviceID := jetsonCamstr(1280, 720, 30, 0)
	host := ":1234"

	// open webcam
	webcam, err = gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// create the mjpeg stream
	stream = mjpeg.NewStream()

	img := gocv.NewMat()
	defer img.Close()

	fmt.Println("Capturing. Point your browser to " + host)

	// start http server
	http.Handle("/mjpeg", stream)

	go func() {
		for {
			if ok := webcam.Read(&img); !ok {
				fmt.Printf("Bad read:\n")
				time.Sleep(1 * time.Second)
				continue
			}
			if img.Empty() {
				log.Println("Empty image")
				continue
			}

			buf, _ := gocv.IMEncode(".jpg", img)
			stream.UpdateJPEG(buf.GetBytes())
			buf.Close()

			time.Sleep(5 * time.Millisecond)
		}
	}()

	log.Fatal(http.ListenAndServe(host, nil))
}

