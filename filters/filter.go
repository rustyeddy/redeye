package filters

import "gocv.io/x/gocv"

type Filter interface {
	Init(inQ chan *gocv.Mat) (outQ chan *gocv.Mat)
	Description() string
	Start()
}

type Flt struct {
	InQ		chan *gocv.Mat
	OutQ	chan *gocv.Mat
	Desc	string
}

func (f Flt) Description() string {
	return f.description
}

func (f Flt) Init(inQ chan *gocv.Mat) (outQ chan *gocv.Mat) {
	f.InQ = inQ
	f.OutQ = make(chan *gocv.Mat)
	return f.OutQ
}

type FilterMap map[string]Filter

var (
	Filters FilterMap = make(map[string]Filter)
)

func (f FilterMap) Add(name string, flt Filter) {
	f[name] = flt
}

func (f FilterMap) Get(name string) (flt Filter, ok bool) {
	flt, ok = f[name]
	return flt, ok
}

func (f FilterMap) List() (names []string) {
	for n, _ := range f {
		names = append(names, n)
	}
	return names
}

