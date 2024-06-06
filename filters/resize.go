package filters

import (
	"fmt"
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
	fltResize Resize;
)

func init() {
	fltResize.description = "Resize image to fixed width, height"
	fltResize.name = "resize"
	Filters.Add(fltResize)
}

func (r Resize) Init(config string) {
	r.X = 2.0
	r.Y = 2.0
	fmt.Printf("resize init: %f - %f\n", r.X, r.Y)
}

func (r Resize) Filter(img *gocv.Mat) *gocv.Mat {
	fmt.Printf("resize Filter: %f - %f\n", r.X, r.Y)	
	gocv.Resize(*img, img, image.Point{}, r.X, r.Y, gocv.InterpolationArea)
	return img
}

