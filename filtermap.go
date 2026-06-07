package redeye

import (
	"net/http"
	"sort"
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

// ServeHTTP returns the filter registry as a JSON array or an HTML fragment,
// depending on whether the request was issued by htmx.
func (f FilterMap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isHTMX(r) {
		entries := make([]filterEntry, 0, len(f))
		for _, flt := range f {
			entries = append(entries, filterEntry{Name: flt.Name(), Desc: flt.Desc()})
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
		renderTemplate(w, "filters.html", entries)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, f.List())
}
