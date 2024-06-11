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

	FrameQ     chan *Frame
	BufferSize int
}

var (
	Running = false
)

func GetCam(deviceID int) (cam *Cam, err error) {
	cam = &Cam{
		DeviceID:   deviceID,
		BufferSize: 10,
	}
	cam.Cap, err = gocv.VideoCaptureDevice(deviceID)
	if err != nil {
		return nil, err
	}

	cam.FrameQ = make(chan *Frame)
	return cam, nil
}

func (cam *Cam) Play() {
	Running = true

	frames := GetFrameBuffers(cam.BufferSize)
	go func() {
		for Running {
			time.Sleep(5 * time.Millisecond)

			frame := frames.Next()
			cam.Cap.Read(frame.Mat)
			if frame.Mat.Empty() {
				continue
			}
			size := frame.Mat.Size()
			if size[0] <= 0 || size[1] <= 0 {
				continue
			}
			cam.FrameQ <- &frame
		}
		close(cam.FrameQ)
	}()
}

func (cam *Cam) Close() {
	cam.Cap.Close()
}
