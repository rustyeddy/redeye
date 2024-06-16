package redeye

import (
	"fmt"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

type MJPEG struct {
	*mjpeg.Stream
}

func NewMJPEG() (m *MJPEG) {
	m = &MJPEG{}
	m.Stream = mjpeg.NewStream()
	return m
}

func (m *MJPEG) Play() chan *Frame {
	frameQ := make(chan *Frame)

	go func() {
		for {
			frame := <-frameQ
			img := frame.Mat

			fmt.Println("got a frame")

			// Can we reuse the buffer?
			buf, _ := gocv.IMEncode(".jpg", *img)
			m.Stream.UpdateJPEG(buf.GetBytes())
			buf.Close()
			fmt.Println("done witho frame")
		}
	}()
	return frameQ
}
