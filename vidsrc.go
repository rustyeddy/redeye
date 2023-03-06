package redeye

import (
	"net/http"
)

type VideoSource interface {
	Play() (imgQ chan []byte) // provide img channel
	Pause()                   // stop img channel
	Filter(flt *Filter)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
