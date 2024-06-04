package filters

import (
	"fmt"
	"log"
	"gocv.io/x/gocv"	
)

type Pipeline struct {
	Filters []Filter
}

var Pipes Pipeline

func (p *Pipeline) AddFilter(fltname string) bool {
	flt, ok := Filters.Get(fltname)
	if !ok {
		log.Println("ERROR - Failed to find filter: ", fltname)
		return ok
	}
	p.Filters = append(p.Filters, flt)
	return ok
}

func (p *Pipeline) Start(imgQ chan *gocv.Mat) (outQ chan *gocv.Mat) {
	fmt.Println(" --------------------------------------- ")
	fmt.Printf("imgQ: %p\n", imgQ)
	inQ := imgQ
	for name, flt := range p.Filters {
		outQ = flt.Process(inQ)
		fmt.Printf("%d - inQ %p -> outQ %p\n", name, inQ, outQ)
		inQ = outQ
	}
	fmt.Printf("outQ: %p\n", outQ)
	fmt.Println(" --------------------------------------- ")
	return outQ
}
