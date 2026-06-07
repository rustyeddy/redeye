// Package resize provides a bilinear-scale filter.
// Import it for its side effect of registering "resize" in redeye.Filters:
//
//	import _ "github.com/rustyeddy/redeye/filters/resize"
package resize

import (
	"encoding/json"
	"image"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

// Resize scales each frame by configurable X and Y factors.
// Pipeline syntax: "resize:0.5" (uniform) or "resize:0.5:0.75" (per-axis).
type Resize struct {
	mu     sync.RWMutex
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Interp int     `json:"interp"`

	redeye.Flt `json:"-"`
}

var fltResize = &Resize{
	Flt: redeye.NewFlt("resize", "scale the image by X and Y factors"),
}

func init() {
	redeye.Filters.Add("resize", fltResize)
}

// Handler returns an http.Handler that accepts JSON to update the scale
// factors at runtime. Register it manually if desired:
//
//	http.Handle("/filters/resize", resize.Handler())
func Handler() http.Handler {
	return fltResize
}

func (r *Resize) Init(config string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.X = 1.0
	r.Y = 1.0
	if config == "" {
		return
	}
	// "factor" → both axes; "x:y" → per axis.
	parts := strings.SplitN(config, ":", 2)
	if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
		r.X = v
		r.Y = v
	} else {
		slog.Warn("resize invalid scale factor", "value", parts[0], "err", err)
	}
	if len(parts) == 2 {
		if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
			r.Y = v
		} else {
			slog.Warn("resize invalid Y factor", "value", parts[1], "err", err)
		}
	}
}

func (r *Resize) Filter(frame *redeye.Frame) *redeye.Frame {
	r.mu.RLock()
	x, y := r.X, r.Y
	r.mu.RUnlock()
	gocv.Resize(*frame.Mat, frame.Mat, image.Point{}, x, y, gocv.InterpolationArea)
	return frame
}

// ServeHTTP accepts a JSON body to update X, Y, and Interp at runtime.
func (res *Resize) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Interp int     `json:"interp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res.mu.Lock()
	res.X = params.X
	res.Y = params.Y
	res.Interp = params.Interp
	x, y, interp := res.X, res.Y, res.Interp
	res.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Interp int     `json:"interp"`
	}{x, y, interp})
}
