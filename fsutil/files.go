// Package fsutil provides atomic file write helpers.
package fsutil

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// WriteFile atomically writes data to filename with perm, replacing any
// existing regular file. It refuses to replace anything else, symlinks
// included: links are never followed. Missing parent directories are created
// with mode 0700. perm is applied exactly, umask does not apply, and Windows
// ignores it.
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	if filename == "" {
		return errors.New("empty filename")
	}

	filename = filepath.Clean(filename)
	if fi, err := os.Lstat(filename); err == nil && !fi.Mode().IsRegular() {
		return fmt.Errorf("%s: not a regular file", filename)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	f, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := f.Name()
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmpName)
		}
	}()

	if _, err = f.Write(data); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err = f.Chmod(perm); err != nil {
			return err
		}
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	// os.Root.Rename requests POSIX replace semantics on Windows, os.Rename does not.
	// A root needs read access to the directory, plain rename does not.
	if root, rerr := os.OpenRoot(filepath.Dir(filename)); rerr == nil {
		err = root.Rename(filepath.Base(tmpName), filepath.Base(filename))
		_ = root.Close()
	} else {
		err = os.Rename(tmpName, filename)
	}
	if err != nil {
		return err
	}
	syncDir(filename)
	return nil
}

// WriteIfChanged writes data only when it differs from the file's current
// content, so a no-op save does not churn the filesystem or its mtime. A
// skipped write leaves the existing mode alone as well.
func WriteIfChanged(filename string, data []byte, perm os.FileMode) error {
	filename = filepath.Clean(filename)
	fi, err := os.Stat(filename)
	if err == nil {
		// Fast path: a size difference means the content differs.
		if fi.Size() != int64(len(data)) {
			return WriteFile(filename, data, perm)
		}
		// Slow path: sizes match, so compare the bytes.
		var existing []byte
		if existing, err = os.ReadFile(filename); err == nil && bytes.Equal(existing, data) {
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return WriteFile(filename, data, perm)
}

// FileExists reports whether path exists and is a regular file. Symlinks
// are followed.
func FileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsRegular()
}

// syncDir fsyncs the directory holding filename so the rename is durable.
// Best effort: the file is already in place, and some filesystems (and
// Windows) cannot sync a directory.
func syncDir(filename string) {
	if runtime.GOOS == "windows" {
		return
	}
	if dir, err := os.Open(filepath.Dir(filename)); err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
}
