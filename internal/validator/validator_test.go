package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidate_MissingFrontmatter(t *testing.T) {
	target := writeTemp(t, "target.md", "no frontmatter here\n")
	v, err := Validate(target, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) == 0 {
		t.Fatal("expected violation for missing frontmatter")
	}
	if v[0].Field != "frontmatter" {
		t.Fatalf("expected frontmatter violation, got %s", v[0].Field)
	}
}

func TestValidate_MissingTitle(t *testing.T) {
	target := writeTemp(t, "target.md", "---\ndraft: true\n---\nbody\n")
	v, err := Validate(target, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) == 0 {
		t.Fatal("expected violation for missing title")
	}
	if v[0].Field != "title" {
		t.Fatalf("expected title violation, got %s", v[0].Field)
	}
}

func TestValidate_CodeBlockCountMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\n---\n```go\na\n```\n```go\nb\n```\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n```go\na\n```\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "codeBlocks" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected code block count mismatch, got %v", v)
	}
}

func TestValidate_CodeBlockContentMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\n---\n```go\na\n```\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n```go\nb\n```\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "codeBlocks" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected code block content mismatch, got %v", v)
	}
}

func TestValidate_InlineCodeCountMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\n---\nuse `foo` and `bar`\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\nuse `foo`\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "inlineCode" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected inline code count mismatch, got %v", v)
	}
}

func TestValidate_InlineCodeContentChanged(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\n---\nuse `foo`\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\nuse `bar`\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "inlineCode" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected inline code content change, got %v", v)
	}
}

func TestValidate_TranslationKeyMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ntranslationKey: abc\n---\nbody\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ntranslationKey: xyz\ndraft: true\n---\nbody\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "translationKey" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected translationKey mismatch, got %v", v)
	}
}

func TestValidate_CJKInTitle(t *testing.T) {
	target := writeTemp(t, "target.md", "---\ntitle: 測試 title\ndraft: true\n---\nbody\n")
	v, err := Validate(target, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("violations: %v", v)
	found := false
	for _, vv := range v {
		if vv.Field == "title" && strings.Contains(vv.Message, "CJK") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected CJK in title violation, got %v", v)
	}
}

func TestValidate_CJKRatioExceeded(t *testing.T) {
	target := writeTemp(t, "target.md", "---\ntitle: test\ndraft: true\n---\n這是中文這是中文這是中文這是中文這是中文\n")
	v, err := Validate(target, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "language" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected CJK ratio violation, got %v", v)
	}
}

func TestValidate_ValidTranslation(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: 原始標題\ntranslationKey: test\nsource_lang: zh-TW\n---\nuse `code` here\n```go\nfunc main() {}\n```\n")
	target := writeTemp(t, "target.md", "---\ntitle: Translated Title\ntranslationKey: test\ndraft: true\nreviewed: false\nsource_lang: zh-TW\ntarget_lang: en\n---\nuse `code` here\n```go\nfunc main() {}\n```\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 0 {
		t.Fatalf("expected no violations, got %v", v)
	}
}

func TestValidate_MissingURL(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\ncheck https://example.com\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\ncheck nothing\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "urls" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing URL violation, got %v", v)
	}
}

func TestValidate_URLsPresent(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\nsource_lang: zh-TW\ntarget_lang: en\n---\ncheck https://example.com\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\nsource_lang: zh-TW\ntarget_lang: en\n---\ncheck https://example.com\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, vv := range v {
		if vv.Field == "urls" {
			t.Fatalf("unexpected URL violation: %v", vv)
		}
	}
}

func TestValidate_SourceLangMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\nsource_lang: zh-TW\ndraft: true\n---\nbody\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\nsource_lang: ja\ndraft: true\n---\nbody\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "language" && strings.Contains(vv.Message, "mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected source_lang mismatch, got %v", v)
	}
}
