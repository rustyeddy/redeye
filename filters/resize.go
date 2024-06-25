package filters

import (
	"encoding/json"
	"fmt"
	"image"
	"net/http"

	"github.com/rustyeddy/redeye"
	"gocv.io/x/gocv"
)

type Resize struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Interp int     `json:"interp"`

	Flt `json:-`
}

var (
	fltResize *Resize = &Resize{
		Flt: Flt{
			name:        "resize",
			description: "resize the give image",
		},
	}
)

func init() {
	Filters.Add("resize", fltResize)
	http.Handle("/filters/resize", fltResize)
}

func (r *Resize) Init(config string) {
	r.X = 1.0
	r.Y = 1.0
}

func (r *Resize) Filter(frame *redeye.Frame) *redeye.Frame {
	gocv.Resize(*frame.Mat, frame.Mat, image.Point{}, r.X, r.Y, gocv.InterpolationArea)
	return frame
}

func (res *Resize) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&res)
	if err != nil {
		fmt.Fprint(w, err)
	}
	fmt.Fprintf(w, "%+v", res)
}
