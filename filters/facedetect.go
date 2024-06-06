package filters

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

type FaceDetector struct {
	XMLFile string
	Flt

	color      color.RGBA
	classifier gocv.CascadeClassifier
}

var (
	faceDetect = new(FaceDetector)
)

func init() {
	Filters.Add("face-detect", faceDetect)
}

func (flt *FaceDetector) Init(config string) {
	flt.description = "Detect faces with XML Cascade"
	flt.color = color.RGBA{0, 0, 255, 0}
	flt.classifier = gocv.NewCascadeClassifier()
	flt.XMLFile = redeye.GetConfig().CascadeFile
	fmt.Printf("XMLFILE: %s\n", flt.XMLFile)
	if !flt.classifier.Load(flt.XMLFile) {
		log.Printf("Error reading cascade file: %v", faceDetect.XMLFile)
		return
	}

}

func (flt *FaceDetector) Filter(img *gocv.Mat) *gocv.Mat {

	// detect faces
	rects := flt.classifier.DetectMultiScale(*img)

	// draw a rectangle around each face on the original image,
	// along with text identifing as "Human"
	for _, r := range rects {
		gocv.Rectangle(img, r, flt.color, 3)

		size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
		pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
		gocv.PutText(img, "Human", pt, gocv.FontHersheyPlain, 1.2, flt.color, 2)
	}

	return img
}
