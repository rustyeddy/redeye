package redeye

import (
	"time"

	"gocv.io/x/gocv"
)

type ImgSrc interface {
	Play() chan *gocv.Mat
	Close()
}

type Cam struct {
	DeviceID int
	Cap      *gocv.VideoCapture
	Running  bool

	FrameQ chan *Frame
}

var (
	Running = false
)

func GetCam(deviceID int) (cam *Cam, err error) {
	cam = &Cam{DeviceID: deviceID}
	cam.Cap, err = gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, err
	}

	cam.FrameQ = make(chan *Frame)
	return cam, nil
}

func (cam *Cam) Play() {
	Running = true
	ringSize := 10
	frames := make([]Frame, ringSize)

	for i := 0; i < ringSize; i++ {
		frames[i] = NewFrame()
	}

	go func() {
		i := 0
		for Running {
			time.Sleep(5 * time.Millisecond)

			frame := &frames[i]
			i++
			if i == ringSize-1 {
				i = 0
			}

			cam.Cap.Read(frame.Mat)
			if frame.Mat.Empty() {
				continue
			}
			size := frame.Mat.Size()
			if size[0] <= 0 || size[1] <= 0 {
				continue
			}
			cam.FrameQ <- frame
		}
		close(cam.FrameQ)
	}()
}

func (cam *Cam) Close() {
	cam.Cap.Close()
}
