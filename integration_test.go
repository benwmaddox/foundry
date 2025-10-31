package foundry

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIntegrationSiteBuild(t *testing.T) {
	base := t.TempDir()
	tmplDir := filepath.Join(base, "templates")
	i18nDir := filepath.Join(base, "i18n")
	contentDir := filepath.Join(base, "content")
	for _, dir := range []string{
		tmplDir,
		filepath.Join(contentDir, "en"),
		filepath.Join(contentDir, "es"),
		i18nDir,
		filepath.Join(base, "dist"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	write := func(path string, data string) {
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	write(filepath.Join(tmplDir, "page.html"), `{{ define "page.html" -}}
<html lang="{{ .Lang }}"><head><title>{{ t "title" }}</title></head>
<body><main>{{ .BodyHTML }}</main><footer>{{ tf "footer" "Default Footer" (index .Extra "BuildTime") }}</footer></body></html>
{{- end }}`)

	write(filepath.Join(i18nDir, "en.json"), `{"title":"Hello","footer":"Built in %s"}`)
	write(filepath.Join(i18nDir, "es.json"), `{"title":"Hola","footer":"Construido en %s"}`)

	write(filepath.Join(contentDir, "en", "welcome.md"), "# Welcome\n\nHello world.")
	write(filepath.Join(contentDir, "es", "welcome.md"), "# Bienvenido\n\nHola mundo.")

	var logBuf bytes.Buffer
	SetStepLogger(log.New(&logBuf, "", 0))
	defer SetStepLogger(nil)

	type pageData struct {
		Title    string
		BodyHTML template.HTML
		Lang     string
		URL      string
		Extra    map[string]any
	}

	loadPages := func(lang string) ([]pageData, error) {
		var pages []pageData
		dir := filepath.Join(contentDir, lang)
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			src, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, err
			}
			bodyHTML, err := MarkdownToHTML(src)
			if err != nil {
				return nil, err
			}
			title := firstHeading(string(src))
			url := strings.TrimSuffix(entry.Name(), ".md") + ".html"
			pages = append(pages, pageData{
				Title:    title,
				BodyHTML: template.HTML(bodyHTML),
				Lang:     lang,
				URL:      url,
				Extra: map[string]any{
					"BuildTime": "test",
				},
			})
		}
		return pages, nil
	}

	buildLang := func(lang string) error {
		return WithStep("build "+lang, func() error {
			trans, err := LoadTranslations(i18nDir, lang)
			if err != nil {
				return err
			}
			funcs := TemplateFuncs(trans)
			tmpl, err := LoadTemplates(filepath.Join(tmplDir, "*.html"), funcs)
			if err != nil {
				return err
			}

			pages, err := loadPages(lang)
			if err != nil {
				return err
			}

			return ForEachParallel(pages, 2, func(p pageData) {
				outPath := filepath.Join(base, "dist", p.Lang, p.URL)
				if err := WriteIfChanged(outPath, mustRender(t, tmpl, "page.html", p)); err != nil {
					panic(err)
				}
			})
		})
	}

	if err := WithStep("build all languages", func() error {
		for _, lang := range []string{"en", "es"} {
			if err := buildLang(lang); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	readDist := func(lang string) string {
		data, err := os.ReadFile(filepath.Join(base, "dist", lang, "welcome.html"))
		if err != nil {
			t.Fatalf("read dist: %v", err)
		}
		return string(data)
	}

	enHTML := readDist("en")
	if !strings.Contains(enHTML, "<title>Hello</title>") {
		t.Fatalf("expected english translation in html: %s", enHTML)
	}
	if !strings.Contains(enHTML, "<p>Hello world.</p>") {
		t.Fatalf("expected markdown conversion in html: %s", enHTML)
	}

	esHTML := readDist("es")
	if !strings.Contains(esHTML, "<title>Hola</title>") {
		t.Fatalf("expected spanish translation in html: %s", esHTML)
	}

	enFooter := "<footer>Built in test</footer>"
	if !strings.Contains(enHTML, enFooter) {
		t.Fatalf("expected formatted footer, got %s", enHTML)
	}

	// Verify WriteIfChanged does not touch unchanged files.
	distFile := filepath.Join(base, "dist", "en", "welcome.html")
	info1, err := os.Stat(distFile)
	if err != nil {
		t.Fatalf("stat dist: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := buildLang("en"); err != nil {
		t.Fatalf("rebuild en: %v", err)
	}
	info2, err := os.Stat(distFile)
	if err != nil {
		t.Fatalf("stat dist second: %v", err)
	}
	if !info1.ModTime().Equal(info2.ModTime()) {
		t.Fatalf("expected mod time unchanged, got %s vs %s", info1.ModTime(), info2.ModTime())
	}

	if out := logBuf.String(); !strings.Contains(out, "start build all languages") || !strings.Contains(out, "done build all languages") {
		t.Fatalf("expected withstep logging, got %q", out)
	}
	if err := buildLang("es"); err != nil {
		t.Fatalf("rebuild es: %v", err)
	}
}

func mustRender(t *testing.T, tmpl *template.Template, name string, data any) []byte {
	t.Helper()
	rendered, err := RenderTemplate(tmpl, name, data)
	if err != nil {
		t.Fatalf("render template: %v", err)
	}
	return rendered
}

func firstHeading(md string) string {
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}
