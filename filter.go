package redeye

// Filter is the core image-processing interface. A frame flows
// ImgSrc → []Filter → ImOut through the pipeline.
type Filter interface {
	Name() string
	Desc() string
	Init(config string)
	Filter(*Frame) *Frame
}

// Flt is an embeddable struct that satisfies the Name() and Desc() methods
// of the Filter interface. Initialise it with NewFlt.
type Flt struct {
	name        string
	description string
}

func (f *Flt) Desc() string { return f.description }
func (f *Flt) Name() string { return f.name }

// NewFlt returns a Flt initialised with name and description.
// Use it when embedding Flt in a filter implementation from another package:
//
//	type MyFilter struct {
//	    redeye.Flt
//	}
//	var myFilter = &MyFilter{Flt: redeye.NewFlt("my-filter", "does something")}
func NewFlt(name, description string) Flt {
	return Flt{name: name, description: description}
}
