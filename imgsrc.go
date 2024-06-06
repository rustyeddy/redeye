package redeye

import (
	"time"

	"gocv.io/x/gocv"
)

var (
	Running = false
)

type ImgSrc interface {
	Play() chan *gocv.Mat
	Close()
}

type Webcam struct {
	DeviceID int
	Cap      *gocv.VideoCapture

	ImgQ	chan *gocv.Mat
}

func GetWebcam(deviceID int) (cam *Webcam, err error) {
	cam = &Webcam{DeviceID: deviceID}
	cam.Cap, err = gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, err
	}

	cam.ImgQ = make(chan *gocv.Mat)
	return cam, nil
}

func (cam *Webcam) Play(img *gocv.Mat) {
	Running = true
	go func() {
		for Running {
			time.Sleep(5 * time.Millisecond)
			cam.Cap.Read(img)
			if img.Empty() {
				continue
			}
			cam.ImgQ <- img
		}
		close(cam.ImgQ)
	}()
}

func (cam *Webcam) Close() {
	cam.Cap.Close()
}
