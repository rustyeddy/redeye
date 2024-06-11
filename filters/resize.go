package filters

import (
	"image"

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
			name: "resize",
		},
	}
)

func init() {
	Filters.Add("resize", fltResize)
}

func (r *Resize) Init(config string) {
	r.description = "Resize image to fixed width, height"
	r.X = 2.0
	r.Y = 2.0
}

func (r *Resize) Filter(img *gocv.Mat) *gocv.Mat {
	gocv.Resize(*img, img, image.Point{}, r.X, r.Y, gocv.InterpolationArea)
	return img
}
