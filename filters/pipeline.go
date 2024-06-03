package filters

import "gocv.io/x/gocv"

type Pipeline struct {
	Filters []*Filter
}

type Pipeline interface {
	Filter(*gocv.Mat) *gocv.Mat
}

var pipelines map[string]Pipeline

func init() {
	pipelines = make(map[string]Pipeline)
	pipelines["resize"] = Resize{X: 4, Y: 4, Interp: 2}
}

func GetPipeline(name string) Pipeline {
	if pipe, ok := pipelines[name]; ok {
		return pipe
	}
	return nil
}
