package redeye

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// Snapper is implemented by any ImgSrc that can save the current frame to disk.
type Snapper interface {
	Snap(file string) error
}

// activeSnapper is set once at startup by SetSnapper; nil means no source is running.
var activeSnapper atomic.Pointer[Snapper]

// SetSnapper registers the active image source as the snap target.
func SetSnapper(s Snapper) {
	if s == nil {
		activeSnapper.Store(nil)
	} else {
		activeSnapper.Store(&s)
	}
}

func getSnapper() Snapper {
	if p := activeSnapper.Load(); p != nil {
		return *p
	}
	return nil
}

var apiRouteMethods = map[string][]string{
	"/api/camera/play":       {http.MethodPost},
	"/api/camera/pause":      {http.MethodPost},
	"/api/camera/config":     {http.MethodGet, http.MethodPut},
	"/api/storage/clips":     {http.MethodGet},
	"/api/storage/snapshots": {http.MethodGet},
	"/api/storage/clip":      {http.MethodPost},
	"/api/storage/snapshot":  {http.MethodPost},
}

func RegisterAPIRoutes(mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/ws", http.HandlerFunc(wsHandler))
	mux.Handle("/api/filters", Filters)
	mux.Handle("/api/camera/snap", postOnly(http.HandlerFunc(snapHandler)))

	for path, methods := range apiRouteMethods {
		path := path
		methods := append([]string(nil), methods...)
		mux.Handle(path, notImplementedHandler(path, methods))
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("api: json encode: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{"status": "ok"})
}

func snapHandler(w http.ResponseWriter, r *http.Request) {
	s := getSnapper()
	if s == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, map[string]string{"error": "no active camera source"})
		return
	}

	file := r.URL.Query().Get("file")
	if file == "" {
		file = fmt.Sprintf("snapshot-%s.jpg", time.Now().Format("20060102-150405"))
	}

	if err := s.Snap(file); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{"file": file})
}

// postOnly wraps a handler to reject non-POST requests with 405.
func postOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func notImplementedHandler(path string, methods []string) http.Handler {
	allowed := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		allowed[method] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := allowed[r.Method]; !ok {
			w.Header().Set("Allow", strings.Join(methods, ", "))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		writeJSON(w, map[string]string{
			"error":  "not implemented",
			"path":   path,
			"method": r.Method,
		})
	})
}
