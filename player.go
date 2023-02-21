package redeye

import (
	"github.com/hybridgroup/mjpeg"
)

type VideoPlayer interface {
	Play()
}

type MJPEGPlayer struct {
	*mjpeg.Stream
}

func NewMJPEGPlayer() *MJPEGPlayer {
	return &MJPEGPlayer{
		Stream: mjpeg.NewStream(),
	}
}

func (player *MJPEGPlayer) Play() chan []byte {
	vidQ := make(chan []byte)

	go func() {
		for {
			select {
			case jpg := <- vidQ:
				player.Stream.UpdateJPEG(jpg)
			}
		}
	}()
	return vidQ
}
