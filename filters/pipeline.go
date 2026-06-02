package filters

import (
	"log"
	"strings"
)

type Pipeline struct {
	Filters []Filter
}

// NewPipeline builds a Pipeline from a colon-separated pipeline string.
// Each filter name may be followed by one or more configuration tokens before
// the next filter name. Those tokens are joined with ":" and passed to the
// filter's Init method. For example:
//
//	"resize:0.5:face-detect"  →  resize.Init("0.5"), faceDetect.Init("")
//	"resize:0.5:0.8:face-detect" →  resize.Init("0.5:0.8"), faceDetect.Init("")
func NewPipeline(pipestr string) *Pipeline {
	pipeline := &Pipeline{}
	if pipestr == "" {
		return pipeline
	}

	var currentFlt Filter
	var configTokens []string

	addCurrent := func() {
		if currentFlt == nil {
			return
		}
		currentFlt.Init(strings.Join(configTokens, ":"))
		pipeline.Filters = append(pipeline.Filters, currentFlt)
		currentFlt = nil
		configTokens = nil
	}

	for _, token := range strings.Split(pipestr, ":") {
		flt, ok := Filters.Get(token)
		if ok {
			addCurrent()
			currentFlt = flt
		} else if currentFlt != nil {
			configTokens = append(configTokens, token)
		} else {
			log.Println("ERROR - Failed to find filter: ", token)
		}
	}
	addCurrent()

	return pipeline
}

func (p *Pipeline) Close() error {
	return nil
}
