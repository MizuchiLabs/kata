package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONMissingFileIsNoOp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	v := struct{ N int }{N: 3}
	if err := LoadJSON(path, &v); err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	if v.N != 3 {
		t.Fatalf("v = %+v, want the caller's value untouched", v)
	}
}

func TestLoadJSONResetsCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	v := struct{ N int }{N: 3}
	if err := LoadJSON(path, &v); err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	if v.N != 0 {
		t.Fatalf("v = %+v, want zero value after corrupt load", v)
	}
	if !FileExists(path) {
		t.Fatal("corrupt file must stay on disk for recovery")
	}
}

func TestSaveJSONWritesAndSkipsUnchanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	if err := SaveJSON(path, map[string]int{"n": 1}, 0o600); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %04o, want 0600", fi.Mode().Perm())
	}

	var got map[string]int
	if err = LoadJSON(path, &got); err != nil {
		t.Fatalf("LoadJSON: %v", err)
	}
	if got["n"] != 1 {
		t.Fatalf("loaded = %v, want the saved content", got)
	}

	// Loosen the mode so a rewrite would be visible.
	if err = os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	if err = SaveJSON(path, map[string]int{"n": 1}, 0o600); err != nil {
		t.Fatalf("SaveJSON: %v", err)
	}
	fi, err = os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Mode().Perm() != 0o644 {
		t.Fatal("identical content was rewritten")
	}
}
