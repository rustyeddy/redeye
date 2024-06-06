package filters

import (
	"log"
	"strings"
)

type Pipeline struct {
	Filters	[]Filter
}

func NewPipeline(pipestr string) *Pipeline {
	pipeline := &Pipeline{}

	flts := strings.Split(pipestr, ":")

	for _, fltname := range flts {
		flt, ok := Filters.Get(fltname)
		if !ok {
			log.Println("ERROR - Failed to find filter: ", fltname)
			return nil
		}
		flt.Init("")
		pipeline.Filters = append(pipeline.Filters, flt)
	}	
	return pipeline
}

type Pipe struct {
	Filter
}

// func NewPipe(flt Filter, inQ chan *gocv.Mat) (p *Pipe) {
// 	return &Pipe{
// 		Filter: flt,
// 	}
// }

// func (p *Pipe) Init(inQ chan *gocv.Mat) (outQ chan *gocv.Mat) {
// 	p.InQ = inQ
// 	p.OutQ = make(chan *gocv.Mat)
// 	return p.OutQ
// }

// func (p *Pipe) Start() {
// 	go func() {
// 		for redeye.Running {
// 			img := <-p.InQ
// 			img = p.Filter.Filter(img)
// 			p.OutQ <- img
// 		}
// 	}()
// }
