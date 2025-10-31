package foundry

import "testing"

func TestMarkdownToHTML(t *testing.T) {
	html, err := MarkdownToHTML([]byte("# Title\n\nHello *world*.\n"))
	if err != nil {
		t.Fatalf("MarkdownToHTML failed: %v", err)
	}
	want := `<h1 id="title">Title</h1>
<p>Hello <em>world</em>.</p>
`
	if string(html) != want {
		t.Fatalf("unexpected html:\n%s", html)
	}
}
