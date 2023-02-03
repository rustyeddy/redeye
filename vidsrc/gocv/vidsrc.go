package vidsrc

import "github.com/rustyeddy/redeye/img"

type Vidsrc interface {
	Play()
	Pause()
	Snapshot()
	PumpVideo(frames <-chan *img.Frame)
}
