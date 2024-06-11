package redeye

import (
	"fmt"
	"log"

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

				if jpg == nil || len(jpg) == 0 {
					log.Println("We have an empty jpg frame")
					continue
				}
				player.Stream.UpdateJPEG(jpg)
			}
		}
	}()
}
