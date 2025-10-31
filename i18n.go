package foundry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Translations represents a map of translation keys to localized strings.
type Translations map[string]string

// LoadTranslations looks for translation data inside dir for the specified
// language. It supports .json, .yaml, and .yml files named <lang>.<ext>.
// Missing files yield an empty translation map.
func LoadTranslations(dir string, lang string) (Translations, error) {
	if lang == "" {
		return nil, errors.New("foundry: lang is empty")
	}
	if dir == "" {
		dir = "."
	}

	candidates := []string{
		filepath.Join(dir, lang+".json"),
		filepath.Join(dir, lang+".yaml"),
		filepath.Join(dir, lang+".yml"),
	}

	var (
		data []byte
		err  error
		path string
	)

	for _, candidate := range candidates {
		data, err = os.ReadFile(candidate)
		if err == nil {
			path = candidate
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("foundry: load translations: %w", err)
		}
	}

	if path == "" {
		return Translations{}, nil
	}

	content := Translations{}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		if err := unmarshalYAML(data, &content); err != nil {
			return nil, fmt.Errorf("foundry: parse translations yaml: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &content); err != nil {
			return nil, fmt.Errorf("foundry: parse translations json: %w", err)
		}
	}

	return content, nil
}

// TemplateFuncs exposes translation helper functions to Go templates. When a
// key is missing the key itself is returned unless a fallback string is
// provided as the first variadic argument.
func TemplateFuncs(t Translations) template.FuncMap {
	copyMap := make(Translations, len(t))
	for k, v := range t {
		copyMap[k] = v
	}

	get := func(key string, fallback ...string) string {
		if val, ok := copyMap[key]; ok {
			return val
		}
		if len(fallback) > 0 {
			return fallback[0]
		}
		return key
	}

	return template.FuncMap{
		"t": func(key string, args ...any) string {
			return formatString(get(key), args...)
		},
		"tf": func(key string, fallback string, args ...any) string {
			return formatString(get(key, fallback), args...)
		},
		"hasTranslation": func(key string) bool {
			_, ok := copyMap[key]
			return ok
		},
	}
}

func formatString(pattern string, args ...any) string {
	if len(args) == 0 {
		return pattern
	}
	return fmt.Sprintf(pattern, args...)
}

func unmarshalYAML(data []byte, out any) error {
	return yaml.Unmarshal(data, out)
}
