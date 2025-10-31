package foundry

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTranslationsJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "en.json")
	if err := os.WriteFile(path, []byte(`{"hello":"Hello","welcome":"Welcome %s"}`), 0o644); err != nil {
		t.Fatalf("write translation: %v", err)
	}

	trans, err := LoadTranslations(dir, "en")
	if err != nil {
		t.Fatalf("LoadTranslations: %v", err)
	}

	if got, want := trans["hello"], "Hello"; got != want {
		t.Fatalf("hello got %q want %q", got, want)
	}
}

func TestLoadTranslationsYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "es.yaml")
	if err := os.WriteFile(path, []byte("hello: Hola\n"), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	trans, err := LoadTranslations(dir, "es")
	if err != nil {
		t.Fatalf("LoadTranslations: %v", err)
	}
	if got, want := trans["hello"], "Hola"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTemplateFuncs(t *testing.T) {
	trans := Translations{
		"hello":   "Hello",
		"welcome": "Welcome %s",
	}
	funcs := TemplateFuncs(trans)

	got := funcs["t"].(func(string, ...any) string)("hello")
	if got != "Hello" {
		t.Fatalf("t hello -> %q", got)
	}

	got = funcs["tf"].(func(string, string, ...any) string)("missing", "Fallback")
	if got != "Fallback" {
		t.Fatalf("tf fallback -> %q", got)
	}

	got = funcs["t"].(func(string, ...any) string)("welcome", "Ben")
	if got != "Welcome Ben" {
		t.Fatalf("t format -> %q", got)
	}

	has := funcs["hasTranslation"].(func(string) bool)("hello")
	if !has {
		t.Fatalf("expected hasTranslation true")
	}
}

func TestWithStep(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	SetStepLogger(logger)
	defer SetStepLogger(nil)

	err := WithStep("example", func() error { return nil })
	if err != nil {
		t.Fatalf("WithStep returned error: %v", err)
	}

	out := buf.String()
	if out == "" || !containsAll(out, []string{"start example", "done example"}) {
		t.Fatalf("unexpected log output: %q", out)
	}
}

func containsAll(out string, needles []string) bool {
	for _, s := range needles {
		if !bytes.Contains([]byte(out), []byte(s)) {
			return false
		}
	}
	return true
}

func TestTranslationsHelpers(t *testing.T) {
	original := Translations{
		"hello": "Hello",
	}
	clone := original.Clone()
	clone["hello"] = "Hola"

	if original["hello"] != "Hello" {
		t.Fatalf("expected clone mutation to not affect original")
	}

	if got := original.Lookup("hello"); got != "Hello" {
		t.Fatalf("Lookup existing -> %q", got)
	}
	if got := original.Lookup("missing", "fallback"); got != "fallback" {
		t.Fatalf("Lookup fallback -> %q", got)
	}
	if got := original.Lookup("missing"); got != "missing" {
		t.Fatalf("Lookup missing -> %q", got)
	}

	if got := original.Format("hello"); got != "Hello" {
		t.Fatalf("Format existing -> %q", got)
	}
	if got := original.Format("Hi %s", "Ben"); got != "Hi Ben" {
		t.Fatalf("Format fallback -> %q", got)
	}
}
