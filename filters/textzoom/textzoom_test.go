package textzoom

import (
	"image"
	"testing"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

// newTestFrame creates an H×W BGR colour frame with all pixels set to mid-grey.
func newTestFrame(t *testing.T, w, h int) (*redeye.Frame, func()) {
	t.Helper()
	m := gocv.NewMatWithSizeFromScalar(
		gocv.NewScalar(128, 128, 128, 0),
		h, w,
		gocv.MatTypeCV8UC3,
	)
	return &redeye.Frame{Mat: &m}, func() { m.Close() }
}

// --- Init ---

func TestInitDefaultZoom(t *testing.T) {
	tz := &TextZoom{}
	tz.Init("")
	if tz.zoom != defaultZoom {
		t.Errorf("want zoom=%.0f, got %.0f", defaultZoom, tz.zoom)
	}
}

func TestInitCustomZoom(t *testing.T) {
	tz := &TextZoom{}
	tz.Init("40")
	if tz.zoom != 40.0 {
		t.Errorf("want zoom=40, got %v", tz.zoom)
	}
}

func TestInitClampedToMax(t *testing.T) {
	tz := &TextZoom{}
	tz.Init("200")
	if tz.zoom != maxZoom {
		t.Errorf("zoom > maxZoom should be clamped to %.0f, got %v", maxZoom, tz.zoom)
	}
}

func TestInitInvalidStringKeepsDefault(t *testing.T) {
	tz := &TextZoom{}
	tz.Init("notanumber")
	if tz.zoom != defaultZoom {
		t.Errorf("invalid config should keep default %.0f, got %v", defaultZoom, tz.zoom)
	}
}

func TestInitZeroOrNegativeKeepsDefault(t *testing.T) {
	for _, v := range []string{"0", "-5"} {
		tz := &TextZoom{}
		tz.Init(v)
		if tz.zoom != defaultZoom {
			t.Errorf("Init(%q): want default %.0f, got %v", v, defaultZoom, tz.zoom)
		}
	}
}

// --- Filter: output dimensions ---

func TestFilterPreservesFrameDimensions(t *testing.T) {
	tz := &TextZoom{zoom: 20}
	frame, cleanup := newTestFrame(t, 640, 480)
	defer cleanup()

	result := tz.Filter(frame)

	if result.Mat.Cols() != 640 || result.Mat.Rows() != 480 {
		t.Errorf("want 640×480, got %d×%d", result.Mat.Cols(), result.Mat.Rows())
	}
}

func TestFilterAt1xPreservesSize(t *testing.T) {
	tz := &TextZoom{zoom: 1}
	frame, cleanup := newTestFrame(t, 320, 240)
	defer cleanup()

	result := tz.Filter(frame)

	if result.Mat.Cols() != 320 || result.Mat.Rows() != 240 {
		t.Errorf("zoom=1: want 320×240, got %d×%d", result.Mat.Cols(), result.Mat.Rows())
	}
}

func TestFilterAt80xPreservesSize(t *testing.T) {
	tz := &TextZoom{zoom: 80}
	frame, cleanup := newTestFrame(t, 1280, 720)
	defer cleanup()

	result := tz.Filter(frame)

	if result.Mat.Cols() != 1280 || result.Mat.Rows() != 720 {
		t.Errorf("zoom=80: want 1280×720, got %d×%d", result.Mat.Cols(), result.Mat.Rows())
	}
}

func TestFilterDoesNotPanicOnEmptyMat(t *testing.T) {
	tz := &TextZoom{zoom: 20}
	m := gocv.NewMat()
	defer m.Close()
	frame := &redeye.Frame{Mat: &m}

	// Should return without panicking.
	result := tz.Filter(frame)
	if result == nil {
		t.Error("Filter returned nil")
	}
}

// --- Helper functions ---

func TestUnsharpMaskPreservesDimensions(t *testing.T) {
	m := gocv.NewMatWithSize(64, 64, gocv.MatTypeCV8UC3)
	defer m.Close()

	result := unsharpMask(m, 1.0, 0.5)
	defer result.Close()

	if result.Cols() != 64 || result.Rows() != 64 {
		t.Errorf("unsharpMask: want 64×64, got %d×%d", result.Cols(), result.Rows())
	}
}

func TestEnhanceContrastPreservesDimensions(t *testing.T) {
	m := gocv.NewMatWithSizeFromScalar(
		gocv.NewScalar(100, 150, 80, 0),
		60, 80, gocv.MatTypeCV8UC3,
	)
	defer m.Close()

	result := enhanceContrast(m)
	defer result.Close()

	if result.Cols() != 80 || result.Rows() != 60 {
		t.Errorf("enhanceContrast: want 80×60, got %d×%d", result.Cols(), result.Rows())
	}
}

func TestIterativeUpscaleReachesTarget(t *testing.T) {
	cases := []struct {
		srcW, srcH, dstW, dstH int
	}{
		{32, 24, 640, 480},   // 20×
		{16, 9, 1280, 720},   // 80×
		{100, 75, 100, 75},   // no-op
		{200, 150, 400, 300}, // 2×
	}

	for _, tc := range cases {
		src := gocv.NewMatWithSize(tc.srcH, tc.srcW, gocv.MatTypeCV8UC3)
		result := iterativeUpscale(src, tc.dstW, tc.dstH)
		src.Close()

		if result.Cols() != tc.dstW || result.Rows() != tc.dstH {
			t.Errorf("upscale %d×%d→%d×%d: got %d×%d",
				tc.srcW, tc.srcH, tc.dstW, tc.dstH,
				result.Cols(), result.Rows())
		}
		result.Close()
	}
}

func TestROIIsExtractedFromCentre(t *testing.T) {
	// Create a frame that is mostly grey but has a red square in the centre.
	// After zoom the centre should dominate the output.
	w, h := 120, 90
	m := gocv.NewMatWithSizeFromScalar(
		gocv.NewScalar(0, 0, 0, 0), h, w, gocv.MatTypeCV8UC3,
	)
	// Paint centre third red (BGR: 0, 0, 255).
	centre := m.Region(image.Rect(w/3, h/3, w*2/3, h*2/3))
	red := gocv.NewMatWithSizeFromScalar(
		gocv.NewScalar(0, 0, 255, 0), centre.Rows(), centre.Cols(), gocv.MatTypeCV8UC3,
	)
	red.CopyTo(&centre)
	red.Close()
	centre.Close()

	frame := &redeye.Frame{Mat: &m}
	tz := &TextZoom{zoom: 3}
	tz.Filter(frame)

	// Sample the middle pixel of the output — should be red (or close to it).
	mid := m.GetVecbAt(h/2, w/2)
	if mid[2] < 200 { // R channel (BGR[2])
		t.Errorf("centre pixel should be predominantly red, got BGR=%v", mid)
	}
	m.Close()
}
