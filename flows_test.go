package foundry

import (
	"html/template"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFlowAssetPipeline(t *testing.T) {
	base := t.TempDir()
	assetSrcDir := filepath.Join(base, "assets-src")
	distDir := filepath.Join(base, "dist")
	if err := EnsureDir(assetSrcDir); err != nil {
		t.Fatalf("ensure asset src dir: %v", err)
	}
	if err := EnsureDir(distDir); err != nil {
		t.Fatalf("ensure dist dir: %v", err)
	}

	srcAsset := filepath.Join(assetSrcDir, "logo.txt")
	if err := os.WriteFile(srcAsset, []byte("version 1"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	trans := Translations{
		"title": "Static Flow",
	}.Clone()

	tmpl, err := template.New("index.html").Funcs(TemplateFuncs(trans)).Parse(`{{ define "index.html" }}<html><head><title>{{ t "title" }}</title></head><body>ok</body></html>{{ end }}`)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	run := func() error {
		return WithStep("build site", func() error {
			if err := CopyFileIfChanged(srcAsset, filepath.Join(distDir, "assets", "logo.txt")); err != nil {
				return err
			}
			out, err := RenderTemplate(tmpl, "index.html", nil)
			if err != nil {
				return err
			}
			return WriteIfChanged(filepath.Join(distDir, "index.html"), out)
		})
	}

	if err := run(); err != nil {
		t.Fatalf("first run: %v", err)
	}

	distAsset := filepath.Join(distDir, "assets", "logo.txt")
	distHTML := filepath.Join(distDir, "index.html")

	assetInfo1, err := os.Stat(distAsset)
	if err != nil {
		t.Fatalf("stat asset: %v", err)
	}
	htmlInfo1, err := os.Stat(distHTML)
	if err != nil {
		t.Fatalf("stat html: %v", err)
	}

	// Ensure reruns remain stable when inputs don't change.
	time.Sleep(10 * time.Millisecond)
	if err := run(); err != nil {
		t.Fatalf("second run: %v", err)
	}
	assetInfo2, err := os.Stat(distAsset)
	if err != nil {
		t.Fatalf("stat asset second: %v", err)
	}
	htmlInfo2, err := os.Stat(distHTML)
	if err != nil {
		t.Fatalf("stat html second: %v", err)
	}

	if !assetInfo1.ModTime().Equal(assetInfo2.ModTime()) {
		t.Fatalf("asset modtime changed without content changes")
	}
	if !htmlInfo1.ModTime().Equal(htmlInfo2.ModTime()) {
		t.Fatalf("html modtime changed without content changes")
	}

	// Change the asset and ensure copy occurs.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(srcAsset, []byte("version 2"), 0o644); err != nil {
		t.Fatalf("mutate asset: %v", err)
	}
	if err := run(); err != nil {
		t.Fatalf("third run: %v", err)
	}
	assetInfo3, err := os.Stat(distAsset)
	if err != nil {
		t.Fatalf("stat asset third: %v", err)
	}
	if !assetInfo3.ModTime().After(assetInfo2.ModTime()) {
		t.Fatalf("expected asset modtime to update after change")
	}
}
