package frontmatter

import (
	"strings"
	"testing"
)

func TestSplitNoFrontmatter(t *testing.T) {
	doc := Split("# Hello\n\nContent here")
	if doc.Frontmatter != "" {
		t.Fatalf("expected empty frontmatter, got %q", doc.Frontmatter)
	}
	if doc.Body != "# Hello\n\nContent here" {
		t.Fatalf("unexpected body: %q", doc.Body)
	}
}

func TestSplitWithFrontmatter(t *testing.T) {
	input := "---\ntitle: Test\ndraft: true\n---\n# Hello"
	doc := Split(input)
	if doc.Metadata.Title != "Test" {
		t.Fatalf("expected title=Test, got %q", doc.Metadata.Title)
	}
	if doc.Metadata.Draft != true {
		t.Fatalf("expected draft=true, got %v", doc.Metadata.Draft)
	}
	if doc.Body != "# Hello" {
		t.Fatalf("unexpected body: %q", doc.Body)
	}
}

func TestInjectProviderMeta(t *testing.T) {
	doc := Split("---\ntitle: Test\n---\n# Hello")

	pm := ProviderMeta{
		Provider: "google",
		Quality:  "machine_draft",
		Reviewed: false,
		Draft:    true,
	}

	result := InjectProviderMeta(doc, pm)

	if !strings.Contains(result, "translation_provider: google") {
		t.Fatalf("expected translation_provider in output: %s", result)
	}
	if !strings.Contains(result, "translation_quality: machine_draft") {
		t.Fatalf("expected translation_quality in output: %s", result)
	}
	if !strings.Contains(result, "draft: true") {
		t.Fatalf("expected draft: true in output: %s", result)
	}
	if !strings.Contains(result, "reviewed: false") {
		t.Fatalf("expected reviewed: false in output: %s", result)
	}
}

func TestInjectProviderMetaPreservesBody(t *testing.T) {
	doc := Split("---\ntitle: Test\n---\n# Hello\n\nContent here")

	pm := ProviderMeta{
		Provider: "deepl",
		Quality:  "machine_draft",
		Reviewed: false,
		Draft:    true,
	}

	result := InjectProviderMeta(doc, pm)

	if !strings.HasSuffix(result, "# Hello\n\nContent here") {
		t.Fatalf("expected body preserved, got: %s", result)
	}
}
