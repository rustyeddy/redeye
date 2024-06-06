package filters

import (
	"gocv.io/x/gocv"	
)

type Filter interface {
	Name()	string
	Desc()	string
	Init(config string)	
	Filter(*gocv.Mat) *gocv.Mat
}

type Flt struct {
	name		string
	description	string
}

func (f Flt) Desc() string {
	return f.description
}

func (f Flt) Name() string {
	return f.name
}


type FilterMap map[string]Filter

var (
	Filters FilterMap = make(map[string]Filter)
)

func (f FilterMap) Add(flt Filter) {
	f[flt.Name()] = flt
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

