package redeye

import (
	_ "embed"
	"net/http"
)

//go:embed static/index.html
var indexHTML []byte

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}
