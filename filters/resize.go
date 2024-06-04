package filters

import (
	"image"

	"gocv.io/x/gocv"
	"github.com/rustyeddy/redeye"
)

type Resize struct {
	X      float64
	Y      float64
	Interp int

	Flt
}

var (
	fltResize Resize;
)

func init() {
	fltResize.FilterDescription.description = "Resize image to fixed width, height"
	fltResize.X = 2.0
	fltResize.Y = 2.0
	Filters.Add("resize", fltResize)
}


func (flt Resize) Filter(img *gocv.Mat) *gocv.Mat {
	gocv.Resize(*img, img, image.Point{}, flt.X, flt.Y, gocv.InterpolationArea)
	return img
}

func (flt Resize) Start() {
	go func() {
		for redeye.Running {
			img = <-flt.InQ
			img = flt.Filter(img)
			flt.OutQ <- img
		}
	}()
}
