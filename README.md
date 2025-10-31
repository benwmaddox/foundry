# Foundry Core

Foundry Core is a small Go library of stable primitives for generating static sites. It focuses on fast full rebuilds, parallel generation, diff-aware file writes, translation helpers, and template ergonomics while leaving site-specific architecture in your hands.

## Requirements

- Go 1.25.3 or newer

## Features

- File primitives: `WriteIfChanged`, `CopyFileIfChanged`, `EnsureDir`
- HTML templating helpers with pluggable `template.FuncMap`
- Markdown rendering via Goldmark with GitHub-flavored extensions
- Safe parallel execution with panic capture
- Translation loaders for JSON/YAML and template helper functions
- Lightweight logging around build steps

## Quick Start

```bash
go get github.com/benwmaddox/foundry
```

```go
package main

import (
	"html/template"
	"path/filepath"

	"github.com/benwmaddox/foundry"
)

func main() {
	_ = foundry.WithStep("build site", func() error {
		translations, err := foundry.LoadTranslations("i18n", "en")
		if err != nil {
			return err
		}
		tmpl, err := foundry.LoadTemplates("templates/*.html", foundry.TemplateFuncs(translations))
		if err != nil {
			return err
		}

		pages := []template.HTML{ /* site-defined data */ }
		return foundry.ForEachParallel(pages, 4, func(body template.HTML) {
			out := filepath.Join("dist", "index.html")
			if err := foundry.WriteIfChanged(out, []byte(body)); err != nil {
				panic(err)
			}
		})
	})
}
```

See `integration_test.go` for a complete, runnable example of a multilingual site build.

## License

MIT © Ben W. Maddox
