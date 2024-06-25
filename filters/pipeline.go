package filters

import (
	"log"
	"strings"
)

type Pipeline struct {
	Filters []Filter
}

func NewPipeline(pipestr string) *Pipeline {
	pipeline := &Pipeline{}
	if pipestr == "" {
		return pipeline
	}

	flts := strings.Split(pipestr, ":")
	for _, fltname := range flts {
		flt, ok := Filters.Get(fltname)
		if !ok {
			log.Println("ERROR - Failed to find filter: ", fltname)
			continue
		}
		flt.Init("")
		pipeline.Filters = append(pipeline.Filters, flt)
	}
	return pipeline
}

func (p *Pipeline) Close() error {
	for _, fltname := range p.Filters {
		_ = fltname
	}
	return nil
}

type Pipe struct {
	Filter
}
