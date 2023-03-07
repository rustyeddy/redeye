package redeye

import (
	"fmt"

	"github.com/hybridgroup/mjpeg"
)

type VideoPlayer interface {
	Play()
}

type MJPEGPlayer struct {
	Device int
	*mjpeg.Stream
}

func NewMJPEGPlayer(dev int) *MJPEGPlayer {
	return &MJPEGPlayer{
		Device: dev,
		Stream: mjpeg.NewStream(),
	}
}

func (player *MJPEGPlayer) URL() string {
	return fmt.Sprintf("/video/%d", player.Device)
}

func (player *MJPEGPlayer) Play(vidQ chan []byte) {
	go func() {
		for {
			select {
			case jpg := <-vidQ:
				player.Stream.UpdateJPEG(jpg)
			}
		}
	}()
}
