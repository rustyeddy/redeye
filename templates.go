package redeye

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"sort"
)

//go:embed templates
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/*.html"))

// isHTMX reports whether the request was issued by htmx (HX-Request header).
// Using this instead of Accept-header parsing avoids false positives from
// regular browser navigation.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// renderTemplate executes a named template and writes it to w.
// Template names are the base filenames of files in templates/ (e.g. "snap.html").
func renderTemplate(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("template render error", "name", name, "err", err)
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

// snapData is passed to templates/snap.html.
type snapData struct {
	File  string // set on success
	Error string // set on failure
}

// configFormData is passed to templates/config.html.
type configFormData struct {
	Config  *Configuration
	Message string // optional feedback after a save
}

// filterEntry is a name/description pair used in filter and pipeline templates.
type filterEntry struct {
	Name string
	Desc string
}

// pipelineViewData is passed to templates/pipeline.html.
type pipelineViewData struct {
	ActiveStr  string          // descriptor string of the active pipeline
	Active     map[string]bool // set of active filter names
	Registered []filterEntry   // all registered filters, sorted by name
}

// buildPipelineViewData assembles a pipelineViewData from the current state.
func buildPipelineViewData() pipelineViewData {
	p := GetPipeline()
	active := make(map[string]bool, len(p.Filters))
	for _, flt := range p.Filters {
		active[flt.Name()] = true
	}
	entries := make([]filterEntry, 0, len(Filters))
	for _, flt := range Filters {
		entries = append(entries, filterEntry{Name: flt.Name(), Desc: flt.Desc()})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return pipelineViewData{ActiveStr: p.Str, Active: active, Registered: entries}
}
