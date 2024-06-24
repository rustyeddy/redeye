package redeye

import (
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

type MJPEG struct {
	*mjpeg.Stream
	opened bool
}

func NewMJPEG() (m *MJPEG) {
	m = &MJPEG{}
	m.Stream = mjpeg.NewStream()
	return m
}

func (m *MJPEG) Play() chan *Frame {
	// need to close frameQ
	frameQ := make(chan *Frame)
	m.opened = true

	go func() {
		for {
			frame := <-frameQ
			img := frame.Mat

			// Can we reuse the buffer?
			buf, _ := gocv.IMEncode(".jpg", *img)
			m.Stream.UpdateJPEG(buf.GetBytes())
			buf.Close()
		}
	}()
	return frameQ
}

func (m *MJPEG) Close() error {
	if m.opened {

		m.opened = false
		return m.Close()
	}
	return nil
}
