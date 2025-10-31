package foundry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndRenderTemplates(t *testing.T) {
	dir := t.TempDir()
	tplPath := filepath.Join(dir, "page.html")
	templateContent := `{{ define "page.html" }}Hello {{ t "name" }}!{{ end }}`
	if err := os.WriteFile(tplPath, []byte(templateContent), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	tmpl, err := LoadTemplates(filepath.Join(dir, "*.html"), TemplateFuncs(Translations{"name": "World"}))
	if err != nil {
		t.Fatalf("LoadTemplates failed: %v", err)
	}

	out, err := RenderTemplate(tmpl, "page.html", nil)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}
	if got, want := string(out), "Hello World!"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
