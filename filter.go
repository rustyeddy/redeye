package redeye

// Filter is the core image-processing interface. A frame flows
// ImgSrc → []Filter → ImOut through the pipeline.
type Filter interface {
	Name() string
	Desc() string
	Init(config string)
	Filter(*Frame) *Frame
}

// ParamDesc describes one runtime-adjustable parameter of a filter.
type ParamDesc struct {
	Key   string  // form field name / SetParam key
	Label string  // human-readable label shown in the UI
	Type  string  // "float" or "int"
	Min   float64
	Max   float64
	Step  float64
	Value float64 // current value
}

// Parametric is an optional interface for filters that expose runtime-adjustable
// parameters. Implement it alongside Filter to get automatic inline UI controls.
type Parametric interface {
	Params() []ParamDesc
	SetParam(key string, value float64) error
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
