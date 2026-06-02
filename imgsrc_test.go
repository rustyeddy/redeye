package redeye

import (
	"testing"
	"time"

	"gocv.io/x/gocv"
)

func newTestImg(t *testing.T) *Img {
	t.Helper()
	m := gocv.NewMat()
	return &Img{frame: &Frame{Mat: &m}}
}

func TestImgIsRunningFalseBeforePlay(t *testing.T) {
	img := newTestImg(t)
	defer img.Close()

	if img.IsRunning() {
		t.Error("IsRunning should be false before Play is called")
	}
}

func TestImgPlayDeliversSingleFrame(t *testing.T) {
	img := newTestImg(t)
	defer img.Close()

	frameQ := img.Play()

	if !img.IsRunning() {
		t.Error("IsRunning should be true after Play before frame is consumed")
	}

	select {
	case f, ok := <-frameQ:
		if !ok {
			t.Fatal("expected a frame; got channel close immediately")
		}
		if f == nil {
			t.Fatal("received nil frame")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for frame")
	}
}

func TestImgIsRunningFalseAfterConsume(t *testing.T) {
	img := newTestImg(t)
	defer img.Close()

	frameQ := img.Play()
	<-frameQ

	if img.IsRunning() {
		t.Error("IsRunning should be false after the frame has been consumed")
	}
}

func TestImgChannelClosedAfterFrame(t *testing.T) {
	img := newTestImg(t)
	defer img.Close()

	frameQ := img.Play()
	<-frameQ // consume the one frame

	_, ok := <-frameQ
	if ok {
		t.Error("second receive should return ok=false (channel is closed)")
	}
}

func TestImgCloseDoesNotPanic(t *testing.T) {
	img := newTestImg(t)
	img.Play()

	if err := img.Close(); err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}
}

func TestImgCloseBeforeConsumeDoesNotPanic(t *testing.T) {
	// Verifies the old bug is gone: Close() used to call close(frameQ) while
	// a goroutine might be blocked sending — now there is no goroutine.
	img := newTestImg(t)
	img.Play()

	// Close without consuming the frame — should not panic.
	img.Close()
}
