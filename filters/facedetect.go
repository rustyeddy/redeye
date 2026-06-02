package filters

import (
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
	loaded     bool
}

var (
	faceDetect *FaceDetector = &FaceDetector{
		Flt: Flt{
			name:        "face-detect",
			description: "Detect faces with XML cascades",
		},
	}
)

func init() {
	Filters.Add("face-detect", faceDetect)
}

func (flt *FaceDetector) Init(config string) {
	flt.description = "Detect faces with XML Cascade"
	flt.color = color.RGBA{0, 0, 255, 0}
	flt.classifier = gocv.NewCascadeClassifier()
	flt.loaded = false
	if config != "" {
		flt.XMLFile = config
	} else {
		flt.XMLFile = redeye.GetConfig().CascadeFile
	}
	if !flt.classifier.Load(flt.XMLFile) {
		log.Printf("Error reading cascade file: %v", flt.XMLFile)
		return
	}
	flt.loaded = true
}

func (flt *FaceDetector) Filter(frame *redeye.Frame) *redeye.Frame {
	if !flt.loaded {
		return frame
	}

	// detect faces
	rects := flt.classifier.DetectMultiScale(*frame.Mat)

	// draw a rectangle around each face on the original image,
	// along with text identifing as "Human"
	for _, r := range rects {
		gocv.Rectangle(frame.Mat, r, flt.color, 3)

		size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
		pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
		gocv.PutText(frame.Mat, "Human", pt, gocv.FontHersheyPlain, 1.2, flt.color, 2)
	}
	return frame
}
