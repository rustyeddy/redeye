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
	"net/http"
	_ "os"
	"time"
	"strconv"

	_ "net/http/pprof"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

type VideoPlayer struct {
	DeviceID interface{}
	webcam   *gocv.VideoCapture
	stream   *mjpeg.Stream
}

func JetsonCamstr(sensorId int, width int, height int, frame int, flip int) string {
	w := strconv.Itoa(width)
	h := strconv.Itoa(height)
	f := strconv.Itoa(frame)
	fl := strconv.Itoa(flip)
	id := strconv.Itoa(sensorId)

	str := "nvarguscamerasrc sensor_id=" + string(id) + " ! video/x-raw(memory:NVMM), width=(int)" + string(w) + ", height=(int)" +
		string(h) + ", framerate=(fraction)" + string(f) + "/1 ! nvvidconv flip-method=" +
		string(fl) + " ! video/x-raw, width=(int)" + string(w) +
		", height=(int)" + string(h) +
		", format=(string)BGRx ! videoconvert ! video/x-raw, format=(string)BGR ! appsink";
	return str
}

func (vs *VideoPlayer) Stream() chan []byte {
	vidQ := make(chan []byte)

	go func() {
		// open webcam
		webcam, err := gocv.OpenVideoCapture(vs.DeviceID)
		if err != nil {
			fmt.Printf("Error opening capture device: %v\n", vs.DeviceID)
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
	return vidQ
}

func (vs *VideoPlayer) Play(vidQ chan []byte) {

	// start http server
	// create the mjpeg stream
	stream := mjpeg.NewStream()
	http.Handle("/mjpeg", stream)

	
	go func() {
		for {
			select {
			case jpg := <- vidQ:
				log.Println()
				stream.UpdateJPEG(jpg)
			}
		}
	}()

	// parse args
	// deviceID := jetsonCamstr(1280, 720, 30, 0)
	host := ":1234"
	fmt.Println("Capturing. Point your browser to " + host)
	log.Fatal(http.ListenAndServe(host, nil))
}


