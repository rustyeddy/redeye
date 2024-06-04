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

	FilterDescription
}

var (
	fltFaceDetect FaceDetector
)

func init() {
	fltFaceDetect.FilterDescription.description = "Detect faces with XML Cascade"
	Filters.Add("face-detect", fltFaceDetect)
}

func (flt FaceDetector) Filter(img *gocv.Mat) *gocv.Mat {

	return img
}

func (flt FaceDetector) Process(vidQ chan *gocv.Mat) (fltQ chan *gocv.Mat) {

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	flt.XMLFile = redeye.GetConfig().CascadeFile
	// // fname := "data/haarcascade_frontalface_default.xml"
	// xmlFile, err := os.ReadFile(redeye.Config.CascadeFile)
	// if err != nil {
	// 	log.Printf("Error reading cascade file: %v", xmlFile)
	// 	return
	// }
	log.Println("FLT XMLFILE: ", flt.XMLFile)
	if !classifier.Load(flt.XMLFile) {
		log.Printf("Error reading cascade file: %v", flt.XMLFile)
		return
	}

	fltQ = make(chan *gocv.Mat)
	fmt.Printf("FLT Q: %+v\n", fltQ)
	for {
		img := <-vidQ
		fmt.Printf("Facedetect image %p\n", img)

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
		fmt.Printf("SENDING IMAGE TO fltQ: %+v", fltQ)
		fltQ <- img
	}
	return fltQ
}
