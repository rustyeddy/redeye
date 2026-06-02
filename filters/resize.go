package filters

import (
	"encoding/json"
	"image"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

type Resize struct {
	mu     sync.RWMutex
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Interp int     `json:"interp"`

	Flt `json:"-"`
}

var (
	fltResize *Resize = &Resize{
		Flt: Flt{
			name:        "resize",
			description: "resize the give image",
		},
	}
)

func init() {
	Filters.Add("resize", fltResize)
	http.Handle("/filters/resize", fltResize)
}

func (r *Resize) Init(config string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.X = 1.0
	r.Y = 1.0
	if config == "" {
		return
	}
	// config may be "factor" (sets both X and Y) or "x:y" (sets each axis).
	parts := strings.SplitN(config, ":", 2)
	if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
		r.X = v
		r.Y = v
	} else {
		log.Printf("resize: invalid scale factor %q: %v", parts[0], err)
	}
	if len(parts) == 2 {
		if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
			r.Y = v
		} else {
			log.Printf("resize: invalid Y scale factor %q: %v", parts[1], err)
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
