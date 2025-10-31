package foundry

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

// LoadTemplates parses all templates matching the provided glob expression.
// The returned template includes any funcs supplied via funcs.
func LoadTemplates(glob string, funcs template.FuncMap) (*template.Template, error) {
	if glob == "" {
		return nil, fmt.Errorf("foundry: template glob is empty")
	}

	files, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("foundry: invalid template glob: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("foundry: no templates matched glob %q", glob)
	}

	root := template.New("foundry")
	if funcs != nil {
		root = root.Funcs(funcs)
	}

	if _, err := root.ParseFiles(files...); err != nil {
		return nil, fmt.Errorf("foundry: parse templates: %w", err)
	}

	return root, nil
}

// RenderTemplate executes the named template with the supplied data and returns
// the rendered bytes.
func RenderTemplate(t *template.Template, name string, data any) ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("foundry: template is nil")
	}
	if name == "" {
		return nil, fmt.Errorf("foundry: template name is empty")
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, fmt.Errorf("foundry: execute template %q: %w", name, err)
	}
	return buf.Bytes(), nil
}
