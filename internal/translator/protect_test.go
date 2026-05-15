package translator

import (
	"strings"
	"testing"
)

func TestProtect_FencedCodeBlocks(t *testing.T) {
	input := "Hello\n```yaml\nkey: value\n```\nWorld"
	protected, tm := Protect(input)
	if strings.Contains(protected, "key: value") {
		t.Fatal("fenced code block not replaced")
	}
	if strings.Contains(protected, "```") {
		t.Fatal("fence markers not replaced")
	}
	if len(tm.tokens) < 1 {
		t.Fatalf("expected at least 1 token, got %d", len(tm.tokens))
	}
}

func TestProtect_InlineCode(t *testing.T) {
	input := "Use `kubectl apply` to deploy"
	protected, tm := Protect(input)
	if strings.Contains(protected, "kubectl apply") {
		t.Fatal("inline code not replaced")
	}
	if len(tm.tokens) < 1 {
		t.Fatalf("expected at least 1 token, got %d", len(tm.tokens))
	}
}

func TestProtect_URLs(t *testing.T) {
	input := "See https://example.com/docs for details"
	protected, tm := Protect(input)
	if strings.Contains(protected, "https://example.com") {
		t.Fatal("URL not replaced")
	}
	if len(tm.tokens) < 1 {
		t.Fatalf("expected at least 1 token, got %d", len(tm.tokens))
	}
}

func TestRestore_Roundtrip(t *testing.T) {
	input := "Hello\n```yaml\nkey: value\n```\nUse `kubectl apply` and see https://example.com\nWorld"
	protected, tm := Protect(input)
	restored, err := Restore(protected, tm)
	if err != nil {
		t.Fatal(err)
	}
	if restored != input {
		t.Fatalf("roundtrip failed:\nexpected: %q\n got: %q", input, restored)
	}
}

func TestRestore_MissingPlaceholder(t *testing.T) {
	_, tm := Protect("```go\ncode\n```")
	_, err := Restore("no placeholders here", tm)
	if err == nil {
		t.Fatal("expected error for missing placeholder")
	}
}

func TestRestore_DuplicatePlaceholder(t *testing.T) {
	protected, tm := Protect("```go\ncode\n```")
	doubled := protected + " and " + protected
	_, err := Restore(doubled, tm)
	if err == nil {
		t.Fatal("expected error for duplicate placeholder")
	}
}

func TestProtect_NoCodeOrURLs(t *testing.T) {
	input := "Just plain text nothing special."
	protected, tm := Protect(input)
	if protected != input {
		t.Fatalf("plain text should be unchanged, got %q", protected)
	}
	if len(tm.tokens) != 0 {
		t.Fatalf("expected 0 tokens, got %d", len(tm.tokens))
	}
}

func TestProtect_MultipleCodeBlocks(t *testing.T) {
	input := "```go\nfmt.Println()\n```\nprose\n```python\nprint()\n```"
	protected, tm := Protect(input)
	if len(tm.tokens) < 2 {
		t.Fatalf("expected at least 2 tokens, got %d", len(tm.tokens))
	}
	restored, err := Restore(protected, tm)
	if err != nil {
		t.Fatal(err)
	}
	if restored != input {
		t.Fatalf("roundtrip failed:\nexpected: %q\n got: %q", input, restored)
	}
}

func TestProtect_CodeBlockWithLanguage(t *testing.T) {
	input := "```go\npackage main\n```"
	protected, tm := Protect(input)
	if len(tm.tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tm.tokens))
	}
	restored, err := Restore(protected, tm)
	if err != nil {
		t.Fatal(err)
	}
	if restored != input {
		t.Fatalf("roundtrip failed:\nexpected: %q\n got: %q", input, restored)
	}
}

func TestProtect_URLInsideCodeBlockNotReplaced(t *testing.T) {
	input := "```go\n// see https://example.com\n```"
	protected, tm := Protect(input)
	if len(tm.tokens) != 1 {
		t.Fatalf("expected 1 token (code block only), got %d", len(tm.tokens))
	}
	restored, err := Restore(protected, tm)
	if err != nil {
		t.Fatal(err)
	}
	if restored != input {
		t.Fatalf("roundtrip failed:\nexpected: %q\n got: %q", input, restored)
	}
}

func TestProtect_EmptyInput(t *testing.T) {
	protected, tm := Protect("")
	if protected != "" {
		t.Fatalf("empty input should stay empty, got %q", protected)
	}
	if len(tm.tokens) != 0 {
		t.Fatalf("expected 0 tokens, got %d", len(tm.tokens))
	}
}

func TestProtect_URLWithTrailingPunctuation(t *testing.T) {
	input := "See (https://example.com/docs)."
	protected, tm := Protect(input)
	if !strings.Contains(protected, "URL_") {
		t.Fatal("URL with parentheses not replaced")
	}
	restored, err := Restore(protected, tm)
	if err != nil {
		t.Fatal(err)
	}
	if restored != input {
		t.Fatalf("roundtrip failed:\nexpected: %q\n got: %q", input, restored)
	}
}

func TestProtect_MixedContentRoundtrip(t *testing.T) {
	input := "Intro\n```yaml\nkey: value\n```\nUse `kubectl` and visit https://example.com\nMore `inline` and https://other.com\nEnd"
	protected, tm := Protect(input)
	restored, err := Restore(protected, tm)
	if err != nil {
		t.Fatal(err)
	}
	if restored != input {
		t.Fatalf("roundtrip failed:\nexpected: %q\n got: %q", input, restored)
	}
}
