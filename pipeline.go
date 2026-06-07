package redeye

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// Pipeline holds an ordered slice of filters to apply to each frame.
type Pipeline struct {
	Filters []Filter
	Str     string // descriptor string used to build this pipeline
}

// activePipeline is swapped atomically so the frame loop always sees the
// latest pipeline without locking.
var activePipeline atomic.Pointer[Pipeline]

// SetPipeline replaces the active pipeline. Safe to call from any goroutine.
func SetPipeline(p *Pipeline) {
	activePipeline.Store(p)
}

// GetPipeline returns the active pipeline. Returns an empty (pass-through)
// pipeline if none has been set.
func GetPipeline() *Pipeline {
	if p := activePipeline.Load(); p != nil {
		return p
	}
	return &Pipeline{}
}

// ToggleFilter adds or removes a single filter by name from the active
// pipeline. It retries with a CAS loop to handle concurrent calls.
// Config parameters are not preserved when a filter is re-added.
func ToggleFilter(name string) (*Pipeline, error) {
	if _, ok := Filters.Get(name); !ok {
		return nil, fmt.Errorf("unknown filter %q", name)
	}
	for {
		old := activePipeline.Load()
		var cur *Pipeline
		if old != nil {
			cur = old
		} else {
			cur = &Pipeline{}
		}

		names := make([]string, 0, len(cur.Filters))
		found := false
		for _, flt := range cur.Filters {
			if flt.Name() == name {
				found = true
				continue // remove
			}
			names = append(names, flt.Name())
		}
		if !found {
			names = append(names, name)
		}

		newP, err := NewPipeline(strings.Join(names, ":"))
		if err != nil {
			return nil, err
		}
		if activePipeline.CompareAndSwap(old, newP) {
			return newP, nil
		}
	}
}

// NewPipeline builds a Pipeline from a colon-separated descriptor string.
// Each filter name may be followed by configuration tokens before the next
// filter name; those tokens are joined with ":" and passed to Init. Examples:
//
//	"resize:0.5:face-detect"     → resize.Init("0.5"),    face-detect.Init("")
//	"resize:0.5:0.8:face-detect" → resize.Init("0.5:0.8"), face-detect.Init("")
//
// Returns an error if any token in a filter-name position is not registered.
func NewPipeline(pipestr string) (*Pipeline, error) {
	p := &Pipeline{Str: pipestr}
	if pipestr == "" {
		return p, nil
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
			return nil, fmt.Errorf("pipeline: unknown filter %q", tok)
		}
	}
	flush()

	return p, nil
}

func (p *Pipeline) Close() error { return nil }
