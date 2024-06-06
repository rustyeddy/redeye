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

	ImgQ chan *gocv.Mat
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

func (cam *Webcam) Play() {
	Running = true
	ringSize := 10
	var imgring [10]gocv.Mat
	for i := 0; i < ringSize; i++ {
		imgring[i] = gocv.NewMat()
	}

	go func() {
		i := 0
		for Running {
			time.Sleep(5 * time.Millisecond)

			img := &imgring[i]
			i++
			if i == ringSize-1 {
				i = 0
			}

			cam.Cap.Read(img)
			if img.Empty() {
				continue
			}
			size := img.Size()
			if size[0] <= 0 || size[1] <= 0 {
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
