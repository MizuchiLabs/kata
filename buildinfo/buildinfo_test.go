package buildinfo

import (
	"runtime"
	"runtime/debug"
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

func TestApplyBuildInfo(t *testing.T) {
	tests := []struct {
		name    string
		bi      *debug.BuildInfo
		wantVer string
		wantCmt string
		wantDt  string
	}{
		{
			name: "go install build reports module version",
			bi: &debug.BuildInfo{Main: debug.Module{
				Version: "v1.2.3",
			}},
			wantVer: "1.2.3",
			wantCmt: "none",
			wantDt:  "unknown",
		},
		{
			name:    "devel version is ignored",
			bi:      &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			wantVer: "dev",
			wantCmt: "none",
			wantDt:  "unknown",
		},
		{
			name: "vcs settings fill commit and date",
			bi: &debug.BuildInfo{Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
				{Key: "vcs.time", Value: "2026-01-02T03:04:05Z"},
			}},
			wantVer: "dev",
			wantCmt: "abc123",
			wantDt:  "2026-01-02T03:04:05Z",
		},
		{
			name: "dirty worktree suffixes the commit",
			bi: &debug.BuildInfo{Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
				{Key: "vcs.modified", Value: "true"},
			}},
			wantVer: "dev",
			wantCmt: "abc123-dirty",
			wantDt:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(v, c, d string) { Version, Commit, Date = v, c, d }(Version, Commit, Date)

			Version, Commit, Date = "dev", "none", "unknown"
			applyBuildInfo(tt.bi)

			if Version != tt.wantVer {
				t.Errorf("Version = %q, want %q", Version, tt.wantVer)
			}
			if Commit != tt.wantCmt {
				t.Errorf("Commit = %q, want %q", Commit, tt.wantCmt)
			}
			if Date != tt.wantDt {
				t.Errorf("Date = %q, want %q", Date, tt.wantDt)
			}
		})
	}
}

func TestApplyBuildInfoKeepsInjectedValues(t *testing.T) {
	defer func(v, c, d string) { Version, Commit, Date = v, c, d }(Version, Commit, Date)

	Version, Commit, Date = "9.9.9", "injected", "injected-date"
	applyBuildInfo(
		&debug.BuildInfo{Main: debug.Module{Version: "v1.0.0"}, Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abc123"},
			{Key: "vcs.time", Value: "2026-01-02T03:04:05Z"},
			{Key: "vcs.modified", Value: "true"},
		}},
	)

	if Version != "9.9.9" || Commit != "injected" || Date != "injected-date" {
		t.Errorf(
			"injected values were overwritten: version=%q commit=%q date=%q",
			Version,
			Commit,
			Date,
		)
	}
}
