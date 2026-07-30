package buildinfo

import (
	"runtime"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	s := String()
	for _, want := range []string{Version, Commit, Date, runtime.Version(), runtime.GOOS, runtime.GOARCH} {
		if !strings.Contains(s, want) {
			t.Errorf("String() = %q, want it to contain %q", s, want)
		}
	}
}

func TestUserAgent(t *testing.T) {
	ua := UserAgent("test")
	if !strings.HasPrefix(ua, "test/") {
		t.Errorf("UserAgent() = %q, want prefix %q", ua, "test/")
	}
	if !strings.Contains(ua, Version) {
		t.Errorf("UserAgent() = %q, want it to contain version %q", ua, Version)
	}
}
