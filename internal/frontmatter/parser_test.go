package frontmatter

import (
	"testing"
)

func TestSplitSections(t *testing.T) {
	body := "# Title\nintro\n## Sub\ncontent\n### Deep\nmore"
	sections := SplitSections(body)
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	if sections[0].Level != 1 || sections[0].Heading != "Title" {
		t.Errorf("section 0: level=%d heading=%q", sections[0].Level, sections[0].Heading)
	}
	if sections[1].Level != 2 || sections[1].Heading != "Sub" {
		t.Errorf("section 1: level=%d heading=%q", sections[1].Level, sections[1].Heading)
	}
	if sections[2].Level != 3 || sections[2].Heading != "Deep" {
		t.Errorf("section 2: level=%d heading=%q", sections[2].Level, sections[2].Heading)
	}
}

func TestSplitSections_NoHeading(t *testing.T) {
	body := "just text\nmore text"
	sections := SplitSections(body)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].Level != 0 || sections[0].Heading != "" {
		t.Errorf("expected level 0 empty heading, got level=%d heading=%q", sections[0].Level, sections[0].Heading)
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	body := "text\n```go\nfmt.Println(\"hi\")\n```\nmore text\n```python\nprint('hi')\n```\nend"
	blocks := ExtractCodeBlocks(body)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
}

func TestExtractCodeBlocks_None(t *testing.T) {
	blocks := ExtractCodeBlocks("no code here")
	if len(blocks) != 0 {
		t.Fatalf("expected 0 blocks, got %d", len(blocks))
	}
}

func TestExtractInlineCode(t *testing.T) {
	spans := ExtractInlineCode("use `foo` and `bar` here")
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}
	if spans[0] != "foo" || spans[1] != "bar" {
		t.Errorf("spans = %v, want [foo bar]", spans)
	}
}

func TestExtractInlineCode_None(t *testing.T) {
	spans := ExtractInlineCode("no inline code")
	if len(spans) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(spans))
	}
}