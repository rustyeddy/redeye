package redeye

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// --- Img.Snap ---

func TestImgSnapWritesFile(t *testing.T) {
	m := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV8UC3)
	img := &Img{frame: &Frame{Mat: &m}}
	defer img.Close()

	out := filepath.Join(t.TempDir(), "snap.jpg")
	require.NoError(t, img.Snap(out))

	info, err := os.Stat(out)
	require.NoError(t, err, "snap file should exist")
	assert.Greater(t, info.Size(), int64(0), "snap file should not be empty")
}

func TestImgSnapBadPath(t *testing.T) {
	m := gocv.NewMatWithSize(10, 10, gocv.MatTypeCV8UC3)
	img := &Img{frame: &Frame{Mat: &m}}
	defer img.Close()

	err := img.Snap("/nonexistent/dir/snap.jpg")
	assert.Error(t, err)
}

// --- GetRTSP ---

func TestGetRTSPRejectsNonRTSPScheme(t *testing.T) {
	cases := []string{
		"http://camera.local/stream",
		"https://camera.local/stream",
		"/dev/video0",
		"",
		"camera.local/stream",
	}
	for _, url := range cases {
		_, err := GetRTSP(url)
		if err == nil {
			t.Errorf("GetRTSP(%q): expected error for non-RTSP URL, got nil", url)
		}
	}
}

func TestGetRTSPAcceptsRTSPScheme(t *testing.T) {
	// Validates the URL-scheme check passes for rtsp:// and rtsps://.
	// The actual connection attempt will fail (no server), but the error
	// must not be the scheme-validation error — it must be an open/connect error.
	for _, url := range []string{"rtsp://localhost/stream", "rtsps://localhost/stream"} {
		_, err := GetRTSP(url)
		if err != nil && err.Error() == "GetRTSP: URL must begin with rtsp:// or rtsps://, got "+`"`+url+`"` {
			t.Errorf("GetRTSP(%q): scheme was rejected but should have passed validation", url)
		}
		// err may be non-nil (no server running), that's expected and fine.
	}
}
