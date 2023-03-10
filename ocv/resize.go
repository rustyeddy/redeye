package ocv

import (
	"image"

	"gocv.io/x/gocv"
)

// func resizeFilter(img *gocv.Mat) *gocv.Mat {
// 	gocv.Resize(*img, img, image.Point{}, 20.0, 20.0, 1)
// 	// image.Point{X: 400, Y: int(newY)},
// 	return img
// }

func resizeFilter(img *gocv.Mat) *gocv.Mat {
	gocv.Resize(*img, img, image.Point{}, 20.0, 20.0, 1)
	// image.Point{X: 400, Y: int(newY)},
	return img
}
