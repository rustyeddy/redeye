// Package textzoom provides a filter that magnifies the centre of a frame
// to make fine print readable from a downward-facing camera.
//
// Pipeline syntax:
//
//	"text-zoom"      — 1× (default)
//	"text-zoom:40"   — 40×
//	"text-zoom:100"  — 100×
//
// At zoom factor Z the filter extracts a (frameW/Z) × (frameH/Z) region from
// the centre of the frame, enhances its contrast, and iteratively upscales it
// with Lanczos resampling and unsharp masking at each step until it fills the
// full frame again.  Colour is preserved throughout.
//
// Import for its side effect:
//
//	import _ "github.com/rustyeddy/redeye/filters/textzoom"
package textzoom

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"sync"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

const (
	defaultZoom    = 1.0
	maxZoom        = 100.0
	claheClipLimit = 2.0
	claheTileSize  = 8
	denoiseSigma   = 0.5 // light pre-denoise: removes sensor noise, keeps edges
	sharpSigma     = 1.0 // unsharp-mask blur radius at each upscale step
	sharpAmount    = 0.5 // unsharp-mask strength  (0 = none, 1 = strong)
)

// TextZoom is the filter implementation.
type TextZoom struct {
	mu sync.RWMutex

	redeye.Flt
	zoom float64
}

var fltTextZoom = &TextZoom{
	Flt:  redeye.NewFlt("text-zoom", "magnify centre of frame to make fine text readable"),
	zoom: defaultZoom,
}

func init() {
	redeye.Filters.Add("text-zoom", fltTextZoom)
}

// Init parses the zoom factor from the pipeline config string.
// Any value outside (0, 100] is clamped or ignored.
func (tz *TextZoom) Init(config string) {
	tz.mu.Lock()
	defer tz.mu.Unlock()

	tz.zoom = defaultZoom
	if config == "" {
		return
	}
	v, err := strconv.ParseFloat(config, 64)
	if err != nil || v <= 0 {
		return
	}
	if v > maxZoom {
		v = maxZoom
	}
	tz.zoom = v
}

// Filter extracts the centre ROI, enhances contrast, iteratively upscales
// with sharpening, and writes the result back into the frame at original size.
func (tz *TextZoom) Filter(frame *redeye.Frame) *redeye.Frame {
	tz.mu.RLock()
	zoom := effectiveZoom(tz.zoom)
	tz.mu.RUnlock()

	src := frame.Mat
	frameW, frameH := src.Cols(), src.Rows()
	if frameW == 0 || frameH == 0 {
		return frame
	}

	// Compute ROI: centre of frame, sized frame ÷ zoom.
	roiW := max(1, int(math.Round(float64(frameW)/zoom)))
	roiH := max(1, int(math.Round(float64(frameH)/zoom)))
	roiX := (frameW - roiW) / 2
	roiY := (frameH - roiH) / 2

	// Region returns a view sharing memory; copy before any processing.
	view := src.Region(image.Rect(roiX, roiY, roiX+roiW, roiY+roiH))
	roi := gocv.NewMat()
	view.CopyTo(&roi)
	view.Close()
	defer roi.Close()

	enhanced := roi
	// // Contrast enhancement in LAB colour space (touches only luminance).
	// enhanced := enhanceContrast(roi)
	// defer enhanced.Close()

	// Light Gaussian denoise before upscaling — removes sensor noise while
	// leaving stroke edges intact. sigma=0.5 is intentionally gentle.
	denoised := gocv.NewMat()
	defer denoised.Close()
	gocv.GaussianBlur(enhanced, &denoised, image.Point{3, 3},
		denoiseSigma, denoiseSigma, gocv.BorderDefault)

	// Iterative upscale: 2× Lanczos + unsharp mask each step, until the image
	// reaches the original frame dimensions.
	result := iterativeUpscale(denoised, frameW, frameH)
	defer result.Close()

	result.CopyTo(src)
	return frame
}

// Params implements redeye.Parametric, exposing zoom as a runtime UI slider.
func (tz *TextZoom) Params() []redeye.ParamDesc {
	tz.mu.RLock()
	defer tz.mu.RUnlock()

	return []redeye.ParamDesc{
		{Key: "zoom", Label: "Zoom", Type: "float", Min: 0, Max: maxZoom, Step: 1, Value: tz.zoom},
	}
}

// SetParam implements redeye.Parametric.
func (tz *TextZoom) SetParam(key string, value float64) error {
	if key != "zoom" {
		return fmt.Errorf("text-zoom: unknown param %q", key)
	}

	tz.mu.Lock()
	defer tz.mu.Unlock()
	tz.zoom = normalizeZoom(value)
	return nil
}

func normalizeZoom(v float64) float64 {
	if v < 0 {
		return defaultZoom
	}
	if v > maxZoom {
		return maxZoom
	}
	return v
}

func effectiveZoom(v float64) float64 {
	if v <= 0 {
		return defaultZoom
	}
	return v
}

// enhanceContrast applies CLAHE to the L channel of the LAB representation,
// leaving hue and saturation untouched, then converts back to BGR.
func enhanceContrast(src gocv.Mat) gocv.Mat {
	lab := gocv.NewMat()
	defer lab.Close()
	gocv.CvtColor(src, &lab, gocv.ColorBGRToLab)

	channels := gocv.Split(lab)
	defer func() {
		for i := range channels {
			channels[i].Close()
		}
	}()

	clahe := gocv.NewCLAHEWithParams(claheClipLimit, image.Point{claheTileSize, claheTileSize})
	defer clahe.Close()

	lEnhanced := gocv.NewMat()
	defer lEnhanced.Close()
	clahe.Apply(channels[0], &lEnhanced)

	merged := gocv.NewMat()
	defer merged.Close()
	gocv.Merge([]gocv.Mat{lEnhanced, channels[1], channels[2]}, &merged)

	result := gocv.NewMat()
	gocv.CvtColor(merged, &result, gocv.ColorLabToBGR)
	return result
}

// iterativeUpscale doubles the image repeatedly (capped at 2× per step to
// keep Lanczos in its optimal operating range) applying unsharp masking after
// each step.  Stops when targetW × targetH is reached.
func iterativeUpscale(src gocv.Mat, targetW, targetH int) gocv.Mat {
	cur := gocv.NewMat()
	src.CopyTo(&cur)

	for cur.Cols() < targetW || cur.Rows() < targetH {
		// Choose the smaller of 2× and whatever gets us exactly to target.
		fw := float64(targetW) / float64(cur.Cols())
		fh := float64(targetH) / float64(cur.Rows())
		step := math.Min(math.Min(fw, fh), 2.0)
		// Rounding can cause one dimension to hit the target while the other
		// remains just short, leaving step at exactly 1.0 with no progress.
		// Break and let the final exact-fit resize below handle it.
		if step <= 1.0 {
			break
		}

		newW := min(int(math.Round(float64(cur.Cols())*step)), targetW)
		newH := min(int(math.Round(float64(cur.Rows())*step)), targetH)

		upscaled := gocv.NewMat()
		gocv.Resize(cur, &upscaled, image.Point{newW, newH}, 0, 0, gocv.InterpolationLanczos4)
		cur.Close()

		sharpened := unsharpMask(upscaled, sharpSigma, sharpAmount)
		upscaled.Close()
		cur = sharpened
	}

	// Final exact-fit pass (usually a no-op or sub-pixel correction).
	if cur.Cols() != targetW || cur.Rows() != targetH {
		fitted := gocv.NewMat()
		gocv.Resize(cur, &fitted, image.Point{targetW, targetH}, 0, 0, gocv.InterpolationLanczos4)
		cur.Close()
		return fitted
	}
	return cur
}

// unsharpMask sharpens src by subtracting a blurred version:
//
//	result = src × (1 + amount) − blur × amount
//
// sigma controls the blur radius; amount controls sharpening strength.
// Coefficients sum to 1 so overall brightness is preserved.
func unsharpMask(src gocv.Mat, sigma, amount float64) gocv.Mat {
	blurred := gocv.NewMat()
	defer blurred.Close()
	// ksize (0,0) means derive kernel size from sigma automatically.
	gocv.GaussianBlur(src, &blurred, image.Point{0, 0}, sigma, sigma, gocv.BorderDefault)

	result := gocv.NewMat()
	gocv.AddWeighted(src, 1.0+amount, blurred, -amount, 0, &result)
	return result
}
