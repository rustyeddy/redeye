package redeye

import (
	"net/http"
)

// FilterMap is a named registry of Filter implementations.
type FilterMap map[string]Filter

// Filters is the process-wide filter registry. Individual filter packages
// register themselves here from their init() functions via Filters.Add.
var Filters FilterMap = make(FilterMap)

func (f FilterMap) Add(name string, flt Filter) {
	f[name] = flt
}

func (f FilterMap) Get(name string) (Filter, bool) {
	flt, ok := f[name]
	return flt, ok
}

func (f FilterMap) List() []string {
	names := make([]string, 0, len(f))
	for n := range f {
		names = append(names, n)
	}
	return names
}

// ServeHTTP returns a JSON array of registered filter names.
func (f FilterMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, f.List())
}
