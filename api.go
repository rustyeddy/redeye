package redeye

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
		slog.Error("json encode error", "err", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{"status": "ok"})
}

func snapHandler(w http.ResponseWriter, r *http.Request) {
	s := getSnapper()
	if s == nil {
		if isHTMX(r) {
			renderTemplate(w, "snap.html", snapData{Error: "no active camera source"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, map[string]string{"error": "no active camera source"})
		return
	}

	// r.FormValue reads from both the URL query string and the form body,
	// so this is backward-compatible with ?file= API callers.
	file := r.FormValue("file")
	if file == "" {
		file = fmt.Sprintf("snapshot-%s.jpg", time.Now().Format("20060102-150405"))
	}

	if err := s.Snap(file); err != nil {
		if isHTMX(r) {
			renderTemplate(w, "snap.html", snapData{Error: err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	if isHTMX(r) {
		renderTemplate(w, "snap.html", snapData{File: file})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{"file": file})
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if isHTMX(r) {
			renderTemplate(w, "config.html", configFormData{Config: Config})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, Config)

	case http.MethodPut:
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err != nil {
				if isHTMX(r) {
					renderTemplate(w, "config.html", configFormData{Config: Config, Message: "Error: " + err.Error()})
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(w, map[string]string{"error": err.Error()})
				return
			}
			applyConfigForm(r, Config)
		} else {
			if err := json.NewDecoder(r.Body).Decode(Config); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(w, map[string]string{"error": err.Error()})
				return
			}
		}
		if isHTMX(r) {
			renderTemplate(w, "config.html", configFormData{Config: Config, Message: "Configuration saved."})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, Config)

	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// applyConfigForm updates the runtime-configurable fields of c from a form submission.
func applyConfigForm(r *http.Request, c *Configuration) {
	c.Pipeline        = r.FormValue("pipeline")
	c.CascadeFile     = r.FormValue("cascade-file")
	c.MQTTBroker      = r.FormValue("broker")
	c.MQTTTopicPrefix = r.FormValue("topic-prefix")
	c.LogFile         = r.FormValue("log")
	c.LogLevel        = r.FormValue("log-level")
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
