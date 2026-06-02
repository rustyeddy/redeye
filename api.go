package redeye

import (
	"encoding/json"
	"net/http"
	"strings"
)

var apiRouteMethods = map[string][]string{
	"/api/camera/play":       {http.MethodPost},
	"/api/camera/pause":      {http.MethodPost},
	"/api/camera/snap":       {http.MethodPost},
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

	mux.HandleFunc("/health", healthHandler)

	for path, methods := range apiRouteMethods {
		path := path
		methods := append([]string(nil), methods...)
		mux.Handle(path, notImplementedHandler(path, methods))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":  "not implemented",
			"path":   path,
			"method": r.Method,
		})
	})
}
