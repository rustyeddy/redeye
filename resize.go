package redeye

import (
	"image"

	"gocv.io/x/gocv"
)

type Resize struct {
	X      float64
	Y      float64
	Interp int
}

func (flt Resize) Filter(img *gocv.Mat) *gocv.Mat {
	gocv.Resize(*img, img, image.Point{}, flt.X, flt.Y, 3)
	return img
}
