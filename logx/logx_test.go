package logx

import (
	"context"
	"log/slog"
	"strings"
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

func TestRedact(t *testing.T) {
	var buf strings.Builder
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: redact}))
	l.Info("login", "user", "alice", "password", "hunter2", "Token", "abc123")

	out := buf.String()
	if strings.Contains(out, "hunter2") || strings.Contains(out, "abc123") {
		t.Errorf("sensitive value leaked: %s", out)
	}
	if !strings.Contains(out, "user=alice") {
		t.Errorf("non-sensitive value dropped: %s", out)
	}
	if !strings.Contains(out, "password=[REDACTED]") || !strings.Contains(out, "Token=[REDACTED]") {
		t.Errorf("redaction marker missing: %s", out)
	}
}

func TestRedactGroup(t *testing.T) {
	var buf strings.Builder
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: redact}))
	l.Info("login", slog.Group("auth", "user", "alice", "token", "s3cr3t"))

	out := buf.String()
	if strings.Contains(out, "s3cr3t") {
		t.Errorf("group value leaked: %s", out)
	}
	if !strings.Contains(out, "user=alice") {
		t.Errorf("non-sensitive group value dropped: %s", out)
	}
}
