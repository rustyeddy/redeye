package redeye

import (
	"log"
	"strings"
)

// Pipeline holds an ordered slice of filters to apply to each frame.
type Pipeline struct {
	Filters []Filter
}

// NewPipeline builds a Pipeline from a colon-separated descriptor string.
// Each filter name may be followed by configuration tokens before the next
// filter name; those tokens are joined with ":" and passed to Init. Examples:
//
//	"resize:0.5:face-detect"     → resize.Init("0.5"),    face-detect.Init("")
//	"resize:0.5:0.8:face-detect" → resize.Init("0.5:0.8"), face-detect.Init("")
func NewPipeline(pipestr string) *Pipeline {
	p := &Pipeline{}
	if pipestr == "" {
		return p
	}

	var current Filter
	var tokens []string

	flush := func() {
		if current == nil {
			return
		}
		current.Init(strings.Join(tokens, ":"))
		p.Filters = append(p.Filters, current)
		current = nil
		tokens = nil
	}

	for _, tok := range strings.Split(pipestr, ":") {
		if flt, ok := Filters.Get(tok); ok {
			flush()
			current = flt
		} else if current != nil {
			tokens = append(tokens, tok)
		} else {
			log.Println("ERROR - unknown filter:", tok)
		}
	}
	flush()

	return p
}

func (p *Pipeline) Close() error { return nil }
