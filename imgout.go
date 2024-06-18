package redeye

import "gocv.io/x/gocv"

type ImOut interface {
	Play() chan *Frame
	Close() error
}

type Window struct {
	*gocv.Window

	running bool
}

func NewWindow(name string) (w *Window) {
	w = &Window{
		Window: gocv.NewWindow(name),
	}
	return w
}

func (w *Window) Play() (outQ chan *Frame) {
	outQ = make(chan *Frame)
	w.running = true

	go func() {
		for w.running {
			f := <-outQ
			w.IMShow(*f.Mat)
			w.WaitKey(10)
		}
	}()

	return outQ
}
