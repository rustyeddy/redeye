package redeye

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"plugin"
)

// LoadPlugin opens a compiled .so plugin file. The plugin's init() is expected
// to call redeye.Filters.Add to register any filters it provides.
// Requires a CGO-enabled, dynamically linked host binary (not compatible with
// -buildmode=static or the cross-compiled RPi static build).
func LoadPlugin(path string) error {
	if _, err := plugin.Open(path); err != nil {
		return fmt.Errorf("open plugin %s: %w", path, err)
	}
	slog.Info("plugin loaded", "path", path)
	return nil
}

// LoadPlugins walks dir and opens every .so file it finds.
// A missing directory is silently ignored (returns 0, nil).
// Returns the count of successfully loaded plugins.
func LoadPlugins(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read plugin dir %s: %w", dir, err)
	}
	var n int
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".so" {
			continue
		}
		if err := LoadPlugin(filepath.Join(dir, e.Name())); err != nil {
			slog.Warn("plugin skipped", "err", err)
			continue
		}
		n++
	}
	return n, nil
}
