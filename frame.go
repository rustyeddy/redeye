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

type FrameBuffers struct {
	frames []Frame
	idx    int
}

func GetFrameBuffers(size int) (frames *FrameBuffers) {
	frames = &FrameBuffers{
		frames: make([]Frame, size),
	}

	for i := 0; i < size; i++ {
		frames.frames[i] = NewFrame()
	}

	return frames
}

func (frames *FrameBuffers) Next() Frame {
	if frames.idx == len(frames.frames)-1 {
		frames.idx = 0
	} else {
		frames.idx++
	}
	return frames.frames[frames.idx]
}
