package logx

import (
	"context"
	"log/slog"
	"testing"
)

func TestInitLevels(t *testing.T) {
	Init(false)
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Init(false): debug level should be disabled")
	}

	Init(true)
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Init(true): debug level should be enabled")
	}
}
