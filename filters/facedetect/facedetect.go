// Package facedetect provides a Haar-cascade face-detection filter.
// Import it for its side effect of registering "face-detect" in redeye.Filters:
//
//	import _ "github.com/rustyeddy/redeye/filters/facedetect"
package facedetect

import (
	"image"
	"image/color"
	"log"
	"os"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

// FaceDetector draws a bounding box around each detected face using a Haar
// cascade classifier and emits a "detection" event on redeye.Events.
type FaceDetector struct {
	XMLFile string
	redeye.Flt

	color      color.RGBA
	classifier gocv.CascadeClassifier
	loaded     bool
}

var faceDetect = &FaceDetector{
	Flt: redeye.NewFlt("face-detect", "Detect faces with XML cascades"),
}

func init() {
	redeye.Filters.Add("face-detect", faceDetect)
}

func (flt *FaceDetector) Init(config string) {
	flt.color = color.RGBA{0, 0, 255, 0}
	flt.classifier = gocv.NewCascadeClassifier()
	flt.loaded = false
	if config != "" {
		flt.XMLFile = config
	} else {
		flt.XMLFile = redeye.GetConfig().CascadeFile
	}
	if _, err := os.Stat(flt.XMLFile); err != nil {
		log.Printf("face-detect: cascade file not found: %s", flt.XMLFile)
		return
	}
	if !flt.classifier.Load(flt.XMLFile) {
		log.Printf("face-detect: failed to load cascade file: %s", flt.XMLFile)
		return
	}
	flt.loaded = true
}

// detFace is the per-face payload inside a "detection" event.
type detFace struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

func (flt *FaceDetector) Filter(frame *redeye.Frame) *redeye.Frame {
	if !flt.loaded {
		return frame
	}

	rects := flt.classifier.DetectMultiScale(*frame.Mat)

	faces := make([]detFace, 0, len(rects))
	for _, r := range rects {
		gocv.Rectangle(frame.Mat, r, flt.color, 3)

		size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
		pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
		gocv.PutText(frame.Mat, "Human", pt, gocv.FontHersheyPlain, 1.2, flt.color, 2)

		faces = append(faces, detFace{X: r.Min.X, Y: r.Min.Y, W: r.Dx(), H: r.Dy()})
	}

	redeye.Events.Publish(redeye.NewEvent("detection", struct {
		Count int       `json:"count"`
		Faces []detFace `json:"faces"`
	}{Count: len(faces), Faces: faces}))

	return frame
}
