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

func TestStructure_HeadingCountMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n## Heading One\n\n## Heading Two\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n## Heading One\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "H2 headings") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected heading count violation, got %v", v)
	}
}

func TestStructure_ListCountMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n- item one\n- item two\n- item three\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n- item one\n- item two\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "unordered list") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected list count violation, got %v", v)
	}
}

func TestStructure_TableCountMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n| A | B |\n|---|---|\n| 1 | 2 |\n| 3 | 4 |\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n| A | B |\n|---|---|\n| 1 | 2 |\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "table rows") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected table count violation, got %v", v)
	}
}

func TestStructure_BlockquoteMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n> note one\n\n> note two\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n> note one\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "blockquotes") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected blockquote violation, got %v", v)
	}
}

func TestStructure_MatchingStructure(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\nsource_lang: zh-TW\ntarget_lang: en\ndraft: true\n---\n## Heading\n\n- item one\n- item two\n\n> note\n\n`code`\n\n```go\nfunc main() {}\n```\n")
	target := writeTemp(t, "target.md", "---\ntitle: Tgt\nsource_lang: zh-TW\ntarget_lang: en\ndraft: true\n---\n## Heading\n\n- item one\n- item two\n\n> note\n\n`code`\n\n```go\nfunc main() {}\n```\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, vv := range v {
		if vv.Field == "structure" {
			t.Fatalf("unexpected structure violation: %v", vv)
		}
	}
}

func TestStructure_ParagraphCountMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\nFirst paragraph here.\n\nSecond paragraph here.\n\nThird paragraph here.\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\nFirst paragraph here.\n\nSecond and third merged together.\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "paragraphs") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected paragraph count violation, got %v", v)
	}
}

func TestStructure_HeadingOrderMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n## Setup Guide\n\n## Configuration Guide\n\n## Testing Guide\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n## Setup Guide\n\n## Testing Guide\n\n## Configuration Guide\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "heading order mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected heading order violation, got %v", v)
	}
}

func TestStructure_HeadingOrderPreserved(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n## 設定方法\n\n## テスト手順\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n## Setup\n\n## Testing Steps\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "heading order") {
			t.Fatalf("unexpected heading order violation for translated headings: %v", vv)
		}
	}
}

func TestStructure_TableColumnCountMismatch(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n| A | B | C |\n|---|---|---|\n| 1 | 2 | 3 |\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n| A | B |\n|---|---|\n| 1 | 2 |\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "columns") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected table column violation, got %v", v)
	}
}

func TestStructure_OmissionHeuristic(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\nThis is a detailed explanation of how the system works with multiple paragraphs covering setup, configuration, testing, deployment, and troubleshooting steps that an engineer would need to follow.\n\nThe second paragraph covers advanced topics including load balancing, circuit breakers, retry policies, and health checks.\n\nA third paragraph discusses monitoring, alerting, and incident response procedures.\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\nIt works.\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "omission" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected omission violation, got %v", v)
	}
}

func TestTone_AbstractOpenerThreshold(t *testing.T) {
	body := strings.Repeat("The system is designed to handle requests.\n", 6)
	target := writeTemp(t, "target.md", "---\ntitle: Test\ndraft: true\n---\n"+body)

	opts := &ValidateOptions{
		ToneChecks: ToneCheckOptions{
			AbstractOpenerThreshold: 3,
		},
	}
	v, err := Validate(target, "", opts)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "tone" && strings.Contains(vv.Message, "abstract openers") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected abstract opener violation, got %v", v)
	}
}

func TestTone_AbstractOpenerUnderThreshold(t *testing.T) {
	body := "The system is designed to handle requests.\nThe user is expected to follow guidelines.\n"
	target := writeTemp(t, "target.md", "---\ntitle: Test\ndraft: true\n---\n"+body)

	opts := &ValidateOptions{
		ToneChecks: ToneCheckOptions{
			AbstractOpenerThreshold: 3,
		},
	}
	v, err := Validate(target, "", opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, vv := range v {
		if vv.Field == "tone" && strings.Contains(vv.Message, "abstract openers") {
			t.Fatalf("unexpected abstract opener violation: %v", vv)
		}
	}
}

func TestTone_AbstractTermOveruse(t *testing.T) {
	body := "The identity of the identity provider affects identity management. Identity is critical.\n"
	target := writeTemp(t, "target.md", "---\ntitle: Test\ndraft: true\n---\n"+body)

	opts := &ValidateOptions{
		ToneChecks: ToneCheckOptions{
			AbstractTerms: []string{"identity"},
		},
	}
	v, err := Validate(target, "", opts)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "tone" && strings.Contains(vv.Message, "abstract term") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected abstract term violation, got %v", v)
	}
}

func TestTone_HeadingDocLikePrefix(t *testing.T) {
	body := "## Overview of the system\n## Introduction to configuration\n## Prerequisites for setup\n"
	target := writeTemp(t, "target.md", "---\ntitle: Test\ndraft: true\n---\n"+body)

	opts := &ValidateOptions{
		ToneChecks: ToneCheckOptions{
			HeadingDocLikePrefixes: []string{"overview", "introduction", "prerequisites"},
		},
	}
	v, err := Validate(target, "", opts)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "tone" && strings.Contains(vv.Message, "doc-like prefix") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected heading phrasing violation, got %v", v)
	}
}

func TestTone_NoViolationsWhenClean(t *testing.T) {
	body := "## How we reduced latency by 40%%\n\nWe replaced the connection pool with a shared channel. The benchmark shows a 40%% improvement.\n"
	target := writeTemp(t, "target.md", "---\ntitle: Test\ndraft: true\n---\n"+body)

	opts := &ValidateOptions{
		ToneChecks: ToneCheckOptions{
			AbstractOpenerThreshold: 3,
			AbstractTerms:           []string{"identity"},
			HeadingDocLikePrefixes:  []string{"overview", "introduction"},
		},
	}
	v, err := Validate(target, "", opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, vv := range v {
		if vv.Field == "tone" {
			t.Fatalf("unexpected tone violation: %v", vv)
		}
	}
}

func TestHeadingComparison_EmDashOnlyDoesNotFail(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n## 效能分析 — 最佳實踐\n\ncontent\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n## Performance Analysis — Best Practices\n\ncontent\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "heading order mismatch") {
			t.Fatalf("unexpected heading order mismatch for em-dash headings: %v", vv)
		}
	}
}

func TestHeadingComparison_InlineCodeInHeadingDoesNotFail(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n## 如何執行 `kubectl set image`\n\ncontent\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n## How to Run `kubectl set image`\n\ncontent\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "heading order mismatch") {
			t.Fatalf("unexpected heading order mismatch for inline-code headings: %v", vv)
		}
	}
}

func TestHeadingComparison_RealHeadingOrderMismatchStillFails(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n## Setup and Configuration\n\n## Testing and Debugging\n\n## Deployment and Monitoring\n\ncontent\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n## Setup and Configuration\n\n## Deployment and Monitoring\n\n## Testing and Debugging\n\ncontent\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "heading order mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected heading order mismatch for reordered headings, got %v", v)
	}
}

func TestHeadingComparison_TranslatedHeadingsNoRegression(t *testing.T) {
	source := writeTemp(t, "source.md", "---\ntitle: src\ndraft: true\n---\n## 安裝指南\n\n## 設定步驟\n\ncontent\n")
	target := writeTemp(t, "target.md", "---\ntitle: tgt\ndraft: true\n---\n## Installation Guide\n\n## Configuration Steps\n\ncontent\n")
	v, err := Validate(target, source, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, vv := range v {
		if vv.Field == "structure" && strings.Contains(vv.Message, "heading order") {
			t.Fatalf("unexpected heading order mismatch for translated headings: %v", vv)
		}
	}
}
