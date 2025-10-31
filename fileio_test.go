package foundry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteIfChanged_NewFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.txt")

	if err := WriteIfChanged(target, []byte("hello")); err != nil {
		t.Fatalf("WriteIfChanged failed: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if got, want := string(data), "hello"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestWriteIfChanged_Idempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "file.txt")

	if err := WriteIfChanged(target, []byte("first")); err != nil {
		t.Fatalf("initial write: %v", err)
	}
	info1, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat1: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := WriteIfChanged(target, []byte("first")); err != nil {
		t.Fatalf("second write: %v", err)
	}
	info2, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat2: %v", err)
	}
	if !info2.ModTime().Equal(info1.ModTime()) {
		t.Fatalf("modification time changed on identical content")
	}
}

func TestCopyFileIfChanged(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "nested", "dst.txt")
	if err := os.WriteFile(src, []byte("copy me"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	if err := CopyFileIfChanged(src, dst); err != nil {
		t.Fatalf("copy: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if got, want := string(data), "copy me"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestEnsureDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a", "b", "c")

	if err := EnsureDir(target); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory, got file")
	}
}
