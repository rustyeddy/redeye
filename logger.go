package redeye

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// InitLogger configures the global slog default logger.
//
// dest controls where logs are written:
//
//	"stderr"         — os.Stderr (default when dest is empty)
//	"stdout"         — os.Stdout
//	any other value  — treated as a file path; the file is created or appended to
//
// level sets the minimum log level: "debug", "info", "warn", or "error".
// Any unrecognised value falls back to "info".
//
// The returned io.Closer is non-nil only when a log file was opened; the
// caller should defer its Close() to flush and release the file handle.
func InitLogger(dest, level string) (io.Closer, error) {
	w, closer, err := openLogDest(dest)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	slog.SetDefault(slog.New(slog.NewTextHandler(w, opts)))

	return closer, nil
}

func openLogDest(dest string) (io.Writer, io.Closer, error) {
	switch strings.ToLower(dest) {
	case "", "stderr":
		return os.Stderr, nil, nil
	case "stdout":
		return os.Stdout, nil, nil
	default:
		f, err := os.OpenFile(dest, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file %q: %w", dest, err)
		}
		return f, f, nil
	}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
