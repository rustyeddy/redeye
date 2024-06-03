package filters

import "gocv.io/x/gocv"

type Filter interface {
	Process(inQ <-chan *gocv.Mat) (outQ chan<- *gocv.Mat)
	Description() string
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

type FilterDescription struct {
	description string
}

func (d FilterDescription) Description() string {
	return d.description
}
