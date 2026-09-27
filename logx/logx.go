// Package logx installs the standard slog configuration: human-readable
// text on a terminal, JSON when piped or running under a service
// manager, always on stderr. Attribute keys in the sensitive set are
// redacted; extend the set with AddSensitiveKeys.
package logx

import (
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
)

const redacted = "[REDACTED]"

// Init sets the default slog logger. debug enables the debug level and
// source locations in log output.
func Init(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level, AddSource: debug, ReplaceAttr: redact}

	var h slog.Handler
	if IsTerminal() {
		h = slog.NewTextHandler(os.Stderr, opts)
	} else {
		h = slog.NewJSONHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(h))
}

// IsTerminal reports whether stderr is a terminal.
func IsTerminal() bool {
	fi, err := os.Stderr.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// sensitiveKeys holds attribute keys whose values must never be logged.
// Lookup is case-insensitive; common spelling variants are listed
// explicitly to avoid per-attribute normalization cost. The default set
// is deliberately minimal: add project-specific keys with AddSensitiveKeys.
var defaultSensitiveKeys = map[string]struct{}{
	"api_key":       {},
	"apikey":        {},
	"authorization": {},
	"auth":          {},
	"cookie":        {},
	"password":      {},
	"passwd":        {},
	"pwd":           {},
	"secret":        {},
	"token":         {},
	"access_token":  {},
	"accesstoken":   {},
	"refresh_token": {},
	"refreshtoken":  {},
	"privkey":       {},
	"privatekey":    {},
	"private_key":   {},
	"cvv":           {},
	"creditcard":    {},
	"credit_card":   {},
	"cardnumber":    {},
	"card_number":   {},
}

// sensitiveKeys is copy-on-write so redact reads it without locking.
var sensitiveKeys atomic.Pointer[map[string]struct{}]

func init() { sensitiveKeys.Store(&defaultSensitiveKeys) }

// AddSensitiveKeys registers keys whose values must be redacted from all
// future log output, matching case-insensitively. Call it before Init.
func AddSensitiveKeys(keys ...string) {
	for {
		old := sensitiveKeys.Load()
		m := maps.Clone(*old)
		for _, k := range keys {
			m[strings.ToLower(k)] = struct{}{}
		}
		if sensitiveKeys.CompareAndSwap(old, &m) {
			return
		}
	}
}

// redact is a slog ReplaceAttr callback that masks the values of
// attributes whose keys appear in sensitiveKeys. Leaf attributes inside
// nested groups are also passed through by the handler, so redaction
// applies recursively. Built-in attributes (time, level, source, msg)
// are passed through unchanged. Error values pass through verbatim;
// panic values are reduced to their type.
func redact(_ []string, a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)

	switch key {
	case "panic":
		return slog.String("panic_type", fmt.Sprintf("%T", a.Value.Any()))
	case "uri":
		a.Value = slog.StringValue(safeURI(a.Value.String()))
	case "endpoint":
		a.Value = slog.StringValue(safeEndpoint(a.Value.String()))
	}

	if _, ok := (*sensitiveKeys.Load())[key]; ok {
		a.Value = slog.StringValue(redacted)
	}
	return a
}

func safeURI(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return redacted
	}
	return u.Scheme + "://" + redacted
}

func safeEndpoint(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Hostname() == "" {
		return redacted
	}
	return u.Scheme + "://" + u.Hostname()
}
