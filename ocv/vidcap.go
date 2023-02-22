package ocv

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
	"log"
	"strconv"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

type CaptureDevice struct {
	DeviceID interface{}
	Pipeline *Pipeline

	*gocv.VideoCapture
}

func GetCaptureDevice(devstr string) *CaptureDevice {
	capdev := &CaptureDevice{
		DeviceID: devstr,
	}

	if strings.HasPrefix(devstr, "rtsp:") {
		cap, err := gocv.OpenVideoCaptureWithAPI(devstr, gocv.VideoCaptureFFmpeg)
		if err != nil {
			log.Printf("Error opening RTSP: %v\n", devstr)
			return nil
		}

		capdev.VideoCapture = cap
		return capdev
	}

	if strings.HasPrefix(devstr, "jetson:") {

		strs := strings.Split(devstr, ":")
		if len(strs) < 2 {
			log.Printf("Error opening capture device: %v\n", devstr)
			return nil
		}
		_, err := strconv.Atoi(strs[1])
		if err != nil {
			log.Printf("Error bad device ID: %v\n", strs[1])
			return nil
		}

		jetstr := JetsonCamstr(strs[1], 1280, 720, 30, 0)
		cap, err := gocv.OpenVideoCapture(jetstr)
		if err != nil {
			log.Printf("Error opening capture device: %v\n", devstr)
			return nil
		}
		capdev.VideoCapture = cap
		return capdev
	}

	log.Printf("Uknown deviceID: %s", devstr)
	return nil
}

func (vc *CaptureDevice) Stream(vidQ chan []byte) {

	go func() {
		img := gocv.NewMat()
		defer img.Close()

		for {
			if ok := vc.VideoCapture.Read(&img); !ok {
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
	}()
}

func JetsonCamstr(sensorId string, width int, height int, frame int, flip int) string {
	w := strconv.Itoa(width)
	h := strconv.Itoa(height)
	f := strconv.Itoa(frame)
	fl := strconv.Itoa(flip)

	str := "nvarguscamerasrc sensor_id=" + sensorId + " ! video/x-raw(memory:NVMM), width=(int)" + string(w) + ", height=(int)" +
		string(h) + ", framerate=(fraction)" + string(f) + "/1 ! nvvidconv flip-method=" +
		string(fl) + " ! video/x-raw, width=(int)" + string(w) +
		", height=(int)" + string(h) +
		", format=(string)BGRx ! videoconvert ! video/x-raw, format=(string)BGR ! appsink"
	return str
}
