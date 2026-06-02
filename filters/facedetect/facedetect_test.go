package facedetect

import (
	"testing"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

func newTestFrame(t *testing.T) (*redeye.Frame, func()) {
	t.Helper()
	m := gocv.NewMat()
	return &redeye.Frame{Mat: &m}, func() { m.Close() }
}

func TestFaceDetectorFilterSafeWhenNeverInitialised(t *testing.T) {
	flt := &FaceDetector{} // loaded=false (zero value)
	frame, cleanup := newTestFrame(t)
	defer cleanup()

	result := flt.Filter(frame)
	if result != frame {
		t.Error("Filter should return the same frame unchanged when classifier is not loaded")
	}
}

func TestFaceDetectorInitBadPathLeavesLoaderFalse(t *testing.T) {
	flt := &FaceDetector{}
	flt.Init("/nonexistent/cascade.xml")

	if flt.loaded {
		t.Error("loaded should be false when the cascade file does not exist")
	}
}

func TestFaceDetectorFilterSafeAfterFailedInit(t *testing.T) {
	flt := &FaceDetector{}
	flt.Init("/nonexistent/cascade.xml")

	frame, cleanup := newTestFrame(t)
	defer cleanup()

	result := flt.Filter(frame)
	if result != frame {
		t.Error("Filter should return frame unchanged after a failed Init")
	}
}

func TestFaceDetectorInitSetsLoadedOnSuccess(t *testing.T) {
	cascadePath := "/usr/local/share/opencv4/haarcascades/haarcascade_frontalface_default.xml"
	flt := &FaceDetector{}
	flt.Init(cascadePath)

	if !flt.loaded {
		t.Skipf("cascade file not available at %s; skipping", cascadePath)
	}
	defer flt.classifier.Close()
}
