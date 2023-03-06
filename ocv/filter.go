package ocv

import (
	"gocv.io/x/gocv"
)

type Pipeline interface {
	Process(imgQ <-chan *gocv.Mat) chan<- *gocv.Mat
	Filter(img *gocv.Mat) *gocv.Mat
}

// Filter interface takes an incoming image, processes it
// then returns a processed image
type Filter struct {
	Name    string
	Filters []*Filter
	Pipeline
}

func (flt *Filter) AddFilters(flts ...*Filter) {
	if flt == nil {
		flt = &Filter{}
	}
	for _, f := range flts {
		f.Filters = append(flt.Filters, f)
	}
}

func (flt *Filter) Filter(img *gocv.Mat) *gocv.Mat {
	return img
}

func (flt *Filter) Process(imgQ <-chan *gocv.Mat) chan<- *gocv.Mat {
	fltQ := make(chan *gocv.Mat)

	for img := range imgQ {
		for _, flt := range flt.Filters {
			img = flt.Filter(img)
		}
		fltQ <- img
	}
	return fltQ
}

type NullFilter struct {
	Filter
}

func (flt *NullFilter) Process(img *gocv.Mat) *gocv.Mat {
	return img
}
