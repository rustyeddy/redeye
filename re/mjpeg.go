package main

import (
	"net/http"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

var (
	mpath string = "/mjpeg"
	mstream *mjpeg.Stream
)

func init() {
	mstream := mjpeg.NewStream()
	http.Handle(mpath, mstream)
}

func updateMJPEG(img gocv.Mat) {
	buf, _ := gocv.IMEncode(".jpg", img)
	defer buf.Close()
	mstream.UpdateJPEG(buf.GetBytes())
}
