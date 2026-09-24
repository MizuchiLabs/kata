package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAndFileExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if !FileExists(path) {
		t.Fatal("FileExists = false after WriteFile")
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", fi.Mode().Perm())
	}
	if FileExists(filepath.Join(t.TempDir(), "missing")) {
		t.Fatal("FileExists = true for a missing file")
	}
}

func TestWriteFileCreatesMissingParents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deep", "dir", "state.json")
	if err := WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if !FileExists(path) {
		t.Fatal("FileExists = false after WriteFile")
	}
	di, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if di.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode = %04o, want 0700", di.Mode().Perm())
	}
}

func TestWriteIfChangedSkipsIdenticalContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := WriteFile(path, []byte("same"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Loosen the mode so a rewrite would be visible.
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	if err := WriteIfChanged(path, []byte("same"), 0o600); err != nil {
		t.Fatalf("WriteIfChanged: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Mode().Perm() != 0o644 {
		t.Fatal("identical content was rewritten")
	}

	// Different content is written.
	if err = WriteIfChanged(path, []byte("changed"), 0o600); err != nil {
		t.Fatalf("WriteIfChanged: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "changed" {
		t.Fatalf("content = %q, want changed", data)
	}
}

func TestWriteFileRefusesNonRegular(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dir")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := WriteFile(path, []byte("x"), 0o600); err == nil {
		t.Fatal("WriteFile must refuse to replace a directory")
	}
}

func TestWriteFileInWriteOnlyDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "drop")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Chmod(dir, 0o333); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer func() { _ = os.Chmod(dir, 0o755) }()

	if err := WriteFile(filepath.Join(dir, "state.json"), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile in a write-only dir: %v", err)
	}
}

func TestWriteFileRefusesSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.json")
	if err := os.WriteFile(target, []byte("original"), 0o600); err != nil {
		t.Fatalf("write target: %v", err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	if err := WriteFile(link, []byte("new"), 0o600); err == nil {
		t.Fatal("WriteFile must refuse to replace a symlink")
	}
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink was replaced")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(data) != "original" {
		t.Fatalf("target content = %q, want original", data)
	}
}
