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

type Cam struct {
	DeviceID int
	Cap      *gocv.VideoCapture

	ImgQ chan *gocv.Mat
}

func GetCam(deviceID int) (cam *Cam, err error) {
	cam = &Cam{DeviceID: deviceID}
	cam.Cap, err = gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, err
	}

	cam.ImgQ = make(chan *gocv.Mat)
	return cam, nil
}

func (cam *Cam) Play(imgring []gocv.Mat) {
	Running = true
	ringSize := len(imgring)
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

func (cam *Cam) Close() {
	cam.Cap.Close()
}
