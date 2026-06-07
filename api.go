package redeye

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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
	mux.HandleFunc("/api/camera/pipeline", pipelineHandler)
	mux.HandleFunc("/api/camera/filter/param", filterParamHandler)
	mux.HandleFunc("/api/plugins/load", pluginLoadHandler)
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
		// Apply the pipeline string immediately so it takes effect without restart.
		if p, err := NewPipeline(Config.Pipeline); err == nil {
			SetPipeline(p)
		} else {
			slog.Warn("config: invalid pipeline string, pipeline unchanged", "err", err)
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

// pipelineHandler manages the active filter pipeline at runtime.
//
//   GET  /api/camera/pipeline          — current pipeline (JSON or HTML)
//   POST /api/camera/pipeline          — set or toggle filters
//
// POST form/JSON fields:
//
//	filter=<name>   toggle that filter on or off in the active pipeline
//	pipeline=<str>  replace the entire pipeline (empty string clears it)
func pipelineHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if isHTMX(r) {
			renderTemplate(w, "pipeline.html", buildPipelineViewData())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]string{"pipeline": GetPipeline().Str})

	case http.MethodPost:
		var filterName, pipeStr string
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			var req struct {
				Filter   string `json:"filter"`
				Pipeline string `json:"pipeline"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(w, map[string]string{"error": err.Error()})
				return
			}
			filterName, pipeStr = req.Filter, req.Pipeline
		} else {
			filterName = r.FormValue("filter")
			pipeStr = r.FormValue("pipeline")
		}

		if filterName != "" {
			if _, err := ToggleFilter(filterName); err != nil {
				if isHTMX(r) {
					renderTemplate(w, "pipeline.html", buildPipelineViewData())
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(w, map[string]string{"error": err.Error()})
				return
			}
		} else {
			p, err := NewPipeline(pipeStr)
			if err != nil {
				if isHTMX(r) {
					renderTemplate(w, "pipeline.html", buildPipelineViewData())
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				writeJSON(w, map[string]string{"error": err.Error()})
				return
			}
			SetPipeline(p)
		}

		if isHTMX(r) {
			renderTemplate(w, "pipeline.html", buildPipelineViewData())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]string{"pipeline": GetPipeline().Str})

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// filterParamHandler updates one or more parameters on an active Parametric filter.
//
//	POST /api/camera/filter/param
//	  filter=<name>   name of the filter to update
//	  <key>=<value>   one form field per parameter (float values)
func filterParamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	filterName := r.FormValue("filter")
	flt, ok := Filters.Get(filterName)
	if !ok {
		http.Error(w, "filter not found: "+filterName, http.StatusNotFound)
		return
	}
	p, ok := flt.(Parametric)
	if !ok {
		http.Error(w, "filter has no parameters: "+filterName, http.StatusBadRequest)
		return
	}
	for key, vals := range r.Form {
		if key == "filter" || len(vals) == 0 {
			continue
		}
		v, err := strconv.ParseFloat(vals[0], 64)
		if err != nil {
			continue
		}
		if err := p.SetParam(key, v); err != nil {
			slog.Warn("filter param set failed", "filter", filterName, "key", key, "err", err)
		}
	}
	if isHTMX(r) {
		renderTemplate(w, "pipeline.html", buildPipelineViewData())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{"ok": "1"})
}

// pluginLoadHandler hot-loads a .so plugin at runtime.
//
//	POST /api/plugins/load
//	  path=<filepath>   path to the .so file (form or JSON {"path":"..."})
//
// On success the pipeline section HTML fragment is returned so htmx can
// refresh the filter list immediately.
func pluginLoadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var path string
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		path = req.Path
	} else {
		path = r.FormValue("path")
	}
	if path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	if err := LoadPlugin(path); err != nil {
		if isHTMX(r) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}
	if isHTMX(r) {
		renderTemplate(w, "pipeline.html", buildPipelineViewData())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, map[string]string{"path": path, "ok": "1"})
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
