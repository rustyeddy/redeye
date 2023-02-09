package main

import (
	"log"

	"gocv.io/x/gocv"
)

func jetsonCamstr() string {
	gstpipe := "nvarguscamerasrc ! " +
		"video/x-raw(memory:NVMM), width=(int)1280, height=(int)720, format=(string)NV12, framerate=(fraction)60/1 ! " +
		"nvvidconv flip-method=0 ! " +
		"video/x-raw, width=(int)1280, height=(int)720, format=(string)BGRx ! " +
		"videoconvert ! " +
		"video/x-raw, format=(string)BGR !" +
		"appsink	"
	return gstpipe
}

func getWindow() *gocv.Window {
	return gocv.NewWindow("Hello")
}

func getImg() gocv.Mat {
	return gocv.NewMat()
}

func streamVideo(camstr string) (frames <-chan gocv.Mat) {

	// Create the channel we are going to pump frames through
	frameQ := make(chan gocv.Mat)

	go func() {

		var cap *gocv.VideoCapture
		log.Println("Opening Video with camstr: ", camstr, "Opening VideoCapture")

		var err error
		cap, err = gocv.OpenVideoCapture(camstr)
		if err != nil {
			log.Fatal("failed to open video capture device")
			return
		}
		defer cap.Close()
		log.Println("Camera streaming  ...")

		window := gocv.NewWindow("Hello")

		// as long as cam.recording is true we will capture images and send
		// them into the image pipeline. We may recieve a REST or MQTT request
		// to stop recording, in that case the cam.recording will be set to
		// false and the recording will stop.
		for true {

			// read a single raw image from the cam.
			// Only a single static image will be in the system at a given time.
			img := gocv.NewMat()

			if ok := cap.Read(&img); !ok {
				log.Println("device closed, turn recording off")
				continue
			}

			// if the image is empty, there will be no sense continueing
			if img.Empty() {
				continue
			}

			window.IMShow(img)
			window.WaitKey(1)

			// buf, _ := gocv.IMEncode(".jpg", img)
			// b := buf.GetBytes()
			// mstream.UpdateJPEG(b)
			// buf.Close()

			// frameQ <- img
		}
		log.Println("Video loop exiting ...")
	}()

	// return the frame channel, our caller will pass it to the reader
	return frameQ
}

