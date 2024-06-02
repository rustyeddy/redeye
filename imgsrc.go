package redeye

import (
	"time"

	"gocv.io/x/gocv"
)

type ImgSrc interface {
	Play() chan *gocv.Mat
	Close()
}

type Webcam struct {
	DeviceID int
	Cap      *gocv.VideoCapture

	running bool
}

func GetWebcam(deviceID int) (cam *Webcam, err error) {
	cam = &Webcam{DeviceID: deviceID}
	cam.Cap, err = gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, err
	}
	return cam, nil
}

func (cam *Webcam) Play() (imgQ chan *gocv.Mat) {

	cam.running = true
	imgQ = make(chan *gocv.Mat)
	img := gocv.NewMat() // This image will be leaked
	go func() {
		for cam.running {
			time.Sleep(5 * time.Millisecond)
			cam.Cap.Read(&img)
			if img.Empty() {
				continue
			}
			imgQ <- &img
		}
		close(imgQ)
	}()
	return imgQ
}

func (cam *Webcam) Close() {
	cam.Cap.Close()
}
