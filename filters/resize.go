package filters

import (
	"image"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

type Resize struct {
	X      float64
	Y      float64
	Interp int
	Flt
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
}

func (r *Resize) Init(config string) {
	r.X = 2.0
	r.Y = 2.0
}

func (r *Resize) Filter(frame *redeye.Frame) *redeye.Frame {
	gocv.Resize(*frame.Mat, frame.Mat, image.Point{}, r.X, r.Y, gocv.InterpolationArea)
	return frame
}
