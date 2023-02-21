package ocv

import (
	"gocv.io/x/gocv"
)

type Filter interface {
	Filter(mat *gocv.Mat) *gocv.Mat
} 

type NullFilter struct {
	Name string
}

func (f NullFilter) Filter(mat *gocv.Mat) *gocv.Mat {
	f.Name = "NullFilter"
	return mat
}

type Pipeline struct {
	Filters []Filter
}

func (p *Pipeline) AddFilters(flts ...Filter) {
	if p == nil {
		p = &Pipeline{}
	}
	for _, f := range flts {
		p.Filters = append(p.Filters, f)		
	}
}

func (p *Pipeline) Pipeline(mat *gocv.Mat) *gocv.Mat {
	m := mat
	for _, flt := range p.Filters {
		m = flt.Filter(m)
	}
	return m
}


