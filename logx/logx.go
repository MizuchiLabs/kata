// Package logx installs the standard slog configuration
// human-readable text on a terminal, JSON when piped or running
// under a service manager, always on stderr.
package logx

import (
	"log/slog"
	"os"
)

// Init sets the default slog logger. debug enables the debug level and
// source locations in log output.
func Init(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level, AddSource: debug}

	var h slog.Handler
	if isTerminal(os.Stderr) {
		h = slog.NewTextHandler(os.Stderr, opts)
	} else {
		h = slog.NewJSONHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(h))
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
