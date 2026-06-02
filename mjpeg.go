package redeye

import (
	"sync"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

type MJPEG struct {
	*mjpeg.Stream
	quit chan struct{}
	wg   sync.WaitGroup
}

func NewMJPEG() (m *MJPEG) {
	m = &MJPEG{
		Stream: mjpeg.NewStream(),
		quit:   make(chan struct{}),
	}
	return m
}

func (m *MJPEG) Play() chan *Frame {
	frameQ := make(chan *Frame)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.quit:
				return
			case frame, ok := <-frameQ:
				if !ok {
					return
				}
				buf, _ := gocv.IMEncode(".jpg", *frame.Mat)
				m.Stream.UpdateJPEG(buf.GetBytes())
				buf.Close()
			}
		}
	}()
	return frameQ
}

func (m *MJPEG) Close() error {
	close(m.quit)
	m.wg.Wait()
	return nil
}
