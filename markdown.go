package foundry

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var markdownEngine = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Strikethrough,
		extension.TaskList,
		extension.Table,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
		html.WithUnsafe(),
	),
)

// MarkdownToHTML converts Markdown bytes to HTML output using a standard,
// CommonMark-compliant renderer with a handful of ergonomic extensions enabled.
func MarkdownToHTML(src []byte) ([]byte, error) {
	if src == nil {
		src = []byte{}
	}

	var buf bytes.Buffer
	if err := markdownEngine.Convert(src, &buf); err != nil {
		return nil, fmt.Errorf("foundry: markdown conversion failed: %w", err)
	}
	return buf.Bytes(), nil
}
