# Foundry Core v1 Development Plan

## 1. Goals

We are building `foundry`:

- A small Go library of **stable primitives** for generating static sites.
- It is NOT a framework: it does not own routing, content models, or build order.
- Every site has its own `main.go` which acts as the build script.
- All real “what pages exist” and “how are they structured” logic lives in site code, not in `foundry`.

We want:

- Fast full regen on every change (`go run build/main.go`)
- Easy per-site customization using normal Go code
- Reuse across many sites in our portfolio
- Optional i18n/translation in templates

## 2. Assumptions

### 2.1 We control all sites using this for now

- We are not designing for random outside users or CMS authors.
- We accept some duplication in per-site code early.

### 2.2 All output is static files on disk

- No runtime server.
- Output dir is `dist/` by convention in each site.

### 2.3 Content model is up to each site

- A site can generate pages from Markdown, from hardcoded Go structs, from external data, etc.
- foundry never assumes “blog”, “page”, “post”, etc. That’s site-level.

### 2.4 We support multiple languages

- We’ll assume localized builds happen per language (e.g. `dist/en/...`, `dist/es/...`).
- Each site is responsible for deciding which languages exist and how data maps to languages.
- foundry just provides translation string lookup for templates.

### 2.5 We expect thousands of pages, not millions

- We want it to be efficient, but we’re not solving petabyte-scale.
- We do care about parallel generation speed and not rewriting unchanged files.

### 2.6 We always run full builds, not partial rebuild/hot-reload

- The build process always runs from scratch as a single pass.
- We keep it fast using diff-aware writes and parallelism.
- We’re fine with simple file watching in dev that just reruns the build binary.

## 3. foundry core v1 API (public surface)

### 3.1 File output primitives

```go
func WriteIfChanged(path string, content []byte) error
func CopyFileIfChanged(srcPath string, dstPath string) error
func EnsureDir(dir string) error
```

### 3.2 Templating

```go
func LoadTemplates(glob string, funcs template.FuncMap) (*template.Template, error)
func RenderTemplate(t *template.Template, name string, data any) ([]byte, error)
```

### 3.3 Markdown

```go
func MarkdownToHTML(src []byte) ([]byte, error)
```

### 3.4 Parallel execution

```go
func ForEachParallel[T any](items []T, workers int, fn func(T)) error
```

### 3.5 i18n helpers

```go
type Translations map[string]string
func LoadTranslations(dir string, lang string) (Translations, error)
func TemplateFuncs(t Translations) template.FuncMap
```

### 3.6 Tiny logging helpers

```go
func WithStep(name string, fn func() error) error
```

## 4. What does NOT go into foundry v1

- No routing rules
- No content discovery
- No opinionated data models
- No sitemap, feed, pagination, blogs, galleries
- No manifest cache across runs yet

## 5. Example site usage

```go
package main

import (
    "os"
    "path/filepath"
    "strings"
    "time"
    "github.com/you/foundry"
    "html/template"
)

type PageData struct {
    Title string
    BodyHTML template.HTML
    Lang  string
    URL   string
    Extra map[string]any
}

func main() {
    _ = foundry.WithStep("build all languages", func() error {
        langs := []string{"en", "es"}

        for _, lang := range langs {
            if err := buildLang(lang); err != nil {
                return err
            }
        }
        return nil
    })
}

func buildLang(lang string) error {
    return foundry.WithStep("build "+lang, func() error {
        trans, _ := foundry.LoadTranslations("i18n", lang)
        funcs := foundry.TemplateFuncs(trans)
        tmpl, _ := foundry.LoadTemplates("templates/*.html", funcs)

        pages := loadPagesForLang(lang)
        return foundry.ForEachParallel(pages, 8, func(p PageData) {
            outPath := filepath.Join("dist", p.Lang, p.URL)
            htmlBytes, _ := foundry.RenderTemplate(tmpl, "page.html", p)
            _ = foundry.WriteIfChanged(outPath, htmlBytes)
        })
    })
}

func loadPagesForLang(lang string) []PageData {
    dir := filepath.Join("content", lang)
    entries, _ := os.ReadDir(dir)

    var out []PageData
    for _, e := range entries {
        if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
            continue
        }
        src, _ := os.ReadFile(filepath.Join(dir, e.Name()))
        bodyHTML, _ := foundry.MarkdownToHTML(src)
        title := firstHeading(string(src))
        url := strings.TrimSuffix(e.Name(), ".md") + ".html"

        out = append(out, PageData{
            Title: title,
            BodyHTML: template.HTML(bodyHTML),
            Lang: lang,
            URL: url,
            Extra: map[string]any{
                "BuildTime": time.Now().UTC().Format(time.RFC3339),
            },
        })
    }
    return out
}

func firstHeading(md string) string {
    for _, line := range strings.Split(md, "\n") {
        if strings.HasPrefix(line, "# ") {
            return strings.TrimSpace(strings.TrimPrefix(line, "# "))
        }
    }
    return ""
}
```

## 6. Development Plan / Milestones

### M1: Core utilities

Implement, test, deliver base functions.

### M2: Reference site

Minimal multilingual site using core functions.

### M3: Integration tests

Functional verification (file writes, parallel correctness, i18n).

### M4: Developer ergonomics

Watch mode, logging, dev flow.

## 7. Integration Tests (Summary)

- i18n substitution correctness
- Parallel page generation correctness
- WriteIfChanged stability
- Dist structure correctness (per-lang)
- Markdown conversion smoke test
- Timing output

## 8. Final Notes

- Strong architecture
- Keep scope limited to primitives
- Promote shared code only after reuse across 3+ sites
