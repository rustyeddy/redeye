package filters

import (
	"fmt"
	"log"
	"strings"
)

type Pipeline struct {
	Filters []Filter
}

func NewPipeline(pipestr string) *Pipeline {
	pipeline := &Pipeline{}

	flts := strings.Split(pipestr, ":")

	for _, fltname := range flts {
		flt, ok := Filters.Get(fltname)
		if !ok {
			log.Println("ERROR - Failed to find filter: ", fltname)
			continue
		}
		flt.Init("")
		pipeline.Filters = append(pipeline.Filters, flt)
		fmt.Printf("NewPipeline: %+v\n", flt)
	}
	return pipeline
}

type Pipe struct {
	Filter
}
