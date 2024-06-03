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

	FilterDescription
}

var (
	fltResize Resize;
)

func init() {
	fltResize.FilterDescription.description = "Resize image to fixed width, height" 
	Filters.Add("resize", fltResize)
}

func (flt Resize) Filter(img *gocv.Mat) *gocv.Mat {
	gocv.Resize(*img, img, image.Point{}, flt.X, flt.Y, 3)
	return img
}

func (flt Resize) Process(inQ chan *gocv.Mat) (outQ chan *gocv.Mat) {
	outQ = make(chan *gocv.Mat)

	var img *gocv.Mat

	go func() {
		for redeye.Running {
			img = <-inQ
			img = flt.Filter(img)
			outQ <- img
		}
	}()

	return outQ
}
