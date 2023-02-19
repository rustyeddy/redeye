package redeye

import (
	"fmt"

	"github.com/hybridgroup/mjpeg"
)

type VideoPlayer interface {
	Play()
}

type MJPEGPlayer struct {
	VidQ chan []byte
	*mjpeg.Stream
}

func NewMJPEGPlayer() *MJPEGPlayer {
	return &MJPEGPlayer{
		Stream: mjpeg.NewStream(),
		VidQ: make(chan []byte),
	}
}

func (player *MJPEGPlayer) Play() {
	go func() {
		for {
			select {
			case jpg := <- player.VidQ:
				player.Stream.UpdateJPEG(jpg)
			}
		}
	}()
}
