package redeye

import (
	"gocv.io/x/gocv"
)

type Frame struct {
	*gocv.Mat
	Meta []byte
}

func NewFrame() (f Frame) {
	f = Frame{}
	m := gocv.NewMat()
	f.Mat = &m

	return f
}
