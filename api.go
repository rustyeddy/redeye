package redeye

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func RegisterAPIRoutes(mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}

	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/ws", http.HandlerFunc(wsHandler))
	mux.Handle("/api/filters", Filters)
	mux.Handle("/api/camera/snap", postOnly(http.HandlerFunc(snapHandler)))
	mux.HandleFunc("/api/camera/config", configHandler)
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

func configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, Config)
	case http.MethodPut:
		if err := json.NewDecoder(r.Body).Decode(Config); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, Config)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
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
