// Package logx installs the standard slog configuration
// human-readable text on a terminal, JSON when piped or running
// under a service manager, always on stderr.
package logx

import (
	"log/slog"
	"os"
	"strings"
)

// Init sets the default slog logger. debug enables the debug level and
// source locations in log output.
func Init(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level, AddSource: debug, ReplaceAttr: redact}

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

// sensitiveKeys holds attribute keys whose values must never be logged.
// Lookup is case-insensitive; common spelling variants are listed
// explicitly to avoid per-attribute normalization cost.
var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"pwd":           {},
	"secret":        {},
	"token":         {},
	"access_token":  {},
	"accesstoken":   {},
	"refresh_token": {},
	"refreshtoken":  {},
	"authorization": {},
	"auth":          {},
	"apikey":        {},
	"api_key":       {},
	"privkey":       {},
	"privatekey":    {},
	"private_key":   {},
	"ssn":           {},
	"creditcard":    {},
	"credit_card":   {},
	"cardnumber":    {},
	"card_number":   {},
	"cvv":           {},
}

const redacted = "[REDACTED]"

// redact is a slog ReplaceAttr callback that masks the values of
// attributes whose keys appear in sensitiveKeys. Leaf attributes inside
// nested groups are also passed through by the handler, so redaction
// applies recursively. Built-in attributes (time, level, source, msg)
// are passed through unchanged.
func redact(_ []string, a slog.Attr) slog.Attr {
	if _, ok := sensitiveKeys[strings.ToLower(a.Key)]; ok {
		a.Value = slog.StringValue(redacted)
	}
	return a
}
