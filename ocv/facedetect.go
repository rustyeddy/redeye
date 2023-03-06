package ocv

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"log"

	"gocv.io/x/gocv"
)

//go:embed data/*.xml
var models embed.FS

type FaceDetector struct {
	XMLFile string
}

func (flt FaceDetector) Filter(vidQ <-chan *gocv.Mat) (fltQ chan<- *gocv.Mat) {

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	fname := "data/haarcascade_frontalface_default.xml"
	xmlFile, err := models.Open(fname)
	if err != nil {
		log.Printf("Error reading cascade file: %v", xmlFile)
		return
	}

	if !classifier.Load(fname) {
		log.Printf("Error reading cascade file: %v", xmlFile)
		return
	}

	fltQ = make(chan *gocv.Mat)

	for {
		img := <-vidQ

		// detect faces
		rects := classifier.DetectMultiScale(*img)
		fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image,
		// along with text identifing as "Human"
		for _, r := range rects {
			gocv.Rectangle(img, r, blue, 3)

			size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}

		// show the image in the window, and wait 1 millisecond
		fltQ <- img
	}
	return fltQ
}
