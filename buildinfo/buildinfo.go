// Package buildinfo exposes version metadata for MizuchiLabs tools.
//
// Release builds inject values via goreleaser ldflags:
//
//	-X github.com/mizuchilabs/kata/buildinfo.Version={{.Version}}
//	-X github.com/mizuchilabs/kata/buildinfo.Commit={{.Commit}}
//	-X github.com/mizuchilabs/kata/buildinfo.Date={{.CommitDate}}
//
// Binaries built without ldflags (go install, local go build) fall back
// to runtime/debug.ReadBuildInfo, so version information is never empty.
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

var (
	// Version is the semantic version without a "v" prefix, e.g. "1.2.3".
	Version = "dev"
	// Commit is the VCS revision the binary was built from.
	Commit = "none"
	// Date is the commit/build timestamp.
	Date = "unknown"
)

func init() {
	if Version != "dev" {
		return // injected via ldflags (release build)
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	applyBuildInfo(bi)
}

// applyBuildInfo fills the package variables from bi. Only variables
// still at their unset defaults ("dev", "none", "unknown") are
// overwritten, so values injected via ldflags always win.
func applyBuildInfo(bi *debug.BuildInfo) {
	if v := bi.Main.Version; v != "" && v != "(devel)" && Version == "dev" {
		Version = strings.TrimPrefix(v, "v") // go install path@vX.Y.Z
	}
	filledCommit := false
	dirty := false
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if Commit == "none" {
				Commit = s.Value
				filledCommit = true
			}
		case "vcs.time":
			if Date == "unknown" {
				Date = s.Value
			}
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if dirty && filledCommit {
		Commit += "-dirty"
	}
}

// String returns the full version string for --version output.
func String() string {
	return fmt.Sprintf("%s (commit %s, built %s, %s %s/%s)",
		Version, Commit, Date, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// UserAgent returns "name/version (goos/arch)" for HTTP clients.
func UserAgent(name string) string {
	return fmt.Sprintf("%s/%s (%s/%s)", name, Version, runtime.GOOS, runtime.GOARCH)
}
