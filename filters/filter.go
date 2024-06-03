package filters

import "gocv.io/x/gocv"

type Filter interface {
	Process(inQ chan *gocv.Mat) (outQ chan *gocv.Mat)
}

type Filters struct {
	filters map[string]Filter
}

func (f *Filters) Add(name string, flt *Filter) {
	filters[name] = flt
}

func (f *Filters) Get(name) (flt *Filter, ok bool) {
	return f[name]
}
