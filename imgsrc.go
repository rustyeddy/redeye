package redeye

import (
	"gocv.io/x/gocv"
)

type ImgSrc interface {
	Play() <-chan *gocv.Mat
}

type Webcam struct {
	DeviceID	int
	Cap			*gocv.VideoCapture
}

func GetWebcam(deviceID int) (cam *Webcam, err error) {
	cam = &Webcam{ DeviceID: deviceID }
	cam.Cap, err = gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, err
	}
	return cam, nil
}

func (cam *Webcam) Play() (imgQ <-chan *gocv.Mat) {
	imgQ = make(chan *gocv.Mat)

	img := gocv.NewMat() 		// This image will be leaked
	go func() {
		cam.Cap.Read(&img)
	}()
	return imgQ
}
