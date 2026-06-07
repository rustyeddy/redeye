//go:build plugin

// Package main is a redeye filter plugin that converts each frame to grayscale.
//
// Build it with:
//
//	go build -buildmode=plugin -o plugins/grayscale.so ./plugins/grayscale
//
// Drop the resulting .so into the plugins/ directory (or whichever path is
// passed via --plugins) and it will be loaded automatically at startup, or
// hot-loaded via POST /api/plugins/load.
//
// Note: the plugin and the host binary must be compiled with the same Go
// toolchain version and identical dependency versions.
package main

import (
	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

type Grayscale struct {
	redeye.Flt
}

func (g *Grayscale) Init(_ string) {}

func (g *Grayscale) Filter(frame *redeye.Frame) *redeye.Frame {
	// Convert to grayscale then back to 3-channel BGR so downstream
	// filters and the MJPEG encoder always see a colour-format frame.
	gocv.CvtColor(*frame.Mat, frame.Mat, gocv.ColorBGRToGray)
	gocv.CvtColor(*frame.Mat, frame.Mat, gocv.ColorGrayToBGR)
	return frame
}

func init() {
	redeye.Filters.Add("grayscale", &Grayscale{
		Flt: redeye.NewFlt("grayscale", "convert frames to grayscale"),
	})
}
