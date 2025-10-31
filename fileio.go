package foundry

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// WriteIfChanged ensures path exists and only writes the provided content to path
// when the bytes differ from the existing file. Writes are performed atomically
// to avoid partially written files when the process is interrupted.
func WriteIfChanged(path string, content []byte) error {
	if path == "" {
		return errors.New("foundry: write path is empty")
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := EnsureDir(dir); err != nil {
			return err
		}
	}

	var existing []byte
	var err error
	existing, err = os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("foundry: read existing file: %w", err)
	}
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	tempFile, err := os.CreateTemp(dirOrDot(dir), ".foundry-*")
	if err != nil {
		return fmt.Errorf("foundry: create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	if _, err := tempFile.Write(content); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("foundry: write temp file: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("foundry: sync temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("foundry: close temp file: %w", err)
	}

	if err := os.Rename(tempFile.Name(), path); err != nil {
		return fmt.Errorf("foundry: rename temp file: %w", err)
	}
	return nil
}

// CopyFileIfChanged copies the file from srcPath to dstPath only if the
// destination content differs from the source.
func CopyFileIfChanged(srcPath string, dstPath string) error {
	if srcPath == "" || dstPath == "" {
		return errors.New("foundry: copy paths must be non-empty")
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("foundry: open source file: %w", err)
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("foundry: stat source file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("foundry: source %s is a directory", srcPath)
	}

	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("foundry: read source file: %w", err)
	}

	return WriteIfChanged(dstPath, data)
}

// EnsureDir creates the directory path (and parents) if it does not exist.
func EnsureDir(dir string) error {
	if dir == "" {
		return errors.New("foundry: directory path is empty")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("foundry: ensure dir: %w", err)
	}
	return nil
}

func dirOrDot(dir string) string {
	if dir == "" || dir == "." {
		return "."
	}
	return dir
}
