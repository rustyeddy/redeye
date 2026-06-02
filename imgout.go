package redeye

import (
	"runtime"
	"sync"

	"gocv.io/x/gocv"
)

type ImOut interface {
	Play() chan *Frame
	Close() error
}

type Window struct {
	*gocv.Window
	quit       chan struct{}
	keyPressed chan struct{}
	wg         sync.WaitGroup
	WaitTime   int
}

func NewWindow(name string) (w *Window) {
	w = &Window{
		Window:     gocv.NewWindow(name),
		quit:       make(chan struct{}),
		keyPressed: make(chan struct{}),
	}
	return w
}

func (w *Window) Close() error {
	close(w.quit)
	w.wg.Wait()
	w.Window.Close()
	return nil
}

// KeyPressed returns a channel that is closed when the user presses a key
// inside the window. Useful for holding the window open on a static image.
func (w *Window) KeyPressed() <-chan struct{} {
	return w.keyPressed
}

func (w *Window) Play() (outQ chan *Frame) {
	outQ = make(chan *Frame)
	w.wg.Add(1)
	go func() {
		// Lock to one OS thread: OpenCV highgui requires all GUI calls on the
		// same thread that created the window.
		runtime.LockOSThread()
		defer w.wg.Done()
		var keyOnce sync.Once
		for {
			select {
			case <-w.quit:
				return
			case f, ok := <-outQ:
				if !ok {
					return
				}
				w.IMShow(*f.Mat)
			}
			// Pump the event loop. For static images (WaitTime==0) poll in
			// 10 ms increments so Close() can interrupt within one tick.
			for {
				key := w.WaitKey(10)
				if key >= 0 {
					keyOnce.Do(func() { close(w.keyPressed) })
					break
				}
				if w.WaitTime > 0 {
					break // camera/video: one tick per frame is enough
				}
				// static image: keep pumping until key press or shutdown
				select {
				case <-w.quit:
					return
				default:
				}
			}
		}
	}()

	return outQ
}
