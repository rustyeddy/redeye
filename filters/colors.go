package filters

import (
	"github.com/rustyeddy/redeye"
)

type ColorDetector struct {
	Flt
}

var (
	colorDetect *ColorDetector = &ColorDetector{
		Flt: Flt{
			name:        "color-detect",
			description: "Detect colors",
		},
	}
)

func init() {
	Filters.Add("color-detect", colorDetect)
}

func (flt *ColorDetector) Init(config string) {

}

func (flt *ColorDetector) Filter(frame *redeye.Frame) *redeye.Frame {

	return frame
}
