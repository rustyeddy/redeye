// Package colors provides a colour-detection filter (stub, not yet implemented).
// Import it for its side effect of registering "color-detect" in redeye.Filters:
//
//	import _ "github.com/rustyeddy/redeye/filters/colors"
package colors

import "github.com/rustyeddy/redeye"

// ColorDetector is a placeholder for a future colour-detection filter.
type ColorDetector struct {
	redeye.Flt
}

var colorDetect = &ColorDetector{
	Flt: redeye.NewFlt("color-detect", "Detect colors (not yet implemented)"),
}

func init() {
	redeye.Filters.Add("color-detect", colorDetect)
}

func (flt *ColorDetector) Init(config string) {}

// Filter is a no-op until the implementation is complete.
func (flt *ColorDetector) Filter(frame *redeye.Frame) *redeye.Frame {
	return frame
}
