package main

import "log"

func main() {

	camstr := jetsonCamstr()
	//camstr := "/dev/video0"
	frameQ := streamVideo(camstr)
	for true {
		select {
		case img := <- frameQ:
			updateMJPEG(img)
		}
	}

	log.Println("Shooting is over")
}
