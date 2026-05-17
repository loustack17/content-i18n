package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
	"github.com/loustack17/content-i18n/internal/core"
)

func setupBatchTest(t *testing.T) (string, *config.Config) {
	t.Helper()
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	enDir := filepath.Join(tmpDir, "en")
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(enDir, 0755)
	os.MkdirAll(promptsDir, 0755)

	os.WriteFile(filepath.Join(promptsDir, "translate-section.md"), []byte("Translate this document."), 0644)

	cfgPath := filepath.Join(tmpDir, "content-i18n.yaml")
	cfgData := `project:
  type: generic-markdown
  source_language: zh-TW
  target_languages:
    - en
paths:
  source: ` + srcDir + `
  targets:
    en: ` + enDir + `
adapter:
  name: generic-markdown
  preserve_relative_paths: true
translation:
  default_provider: ai-harness
  output:
    draft: true
    reviewed: false
    preserve_code_blocks: true
    preserve_inline_code: true
    preserve_links: true
style:
  pack: ""
  glossary: ""
`
	os.WriteFile(cfgPath, []byte(cfgData), 0644)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	return tmpDir, cfg
}

func writeSourceFile(t *testing.T, path string) {
	t.Helper()
	os.WriteFile(path, []byte(`---
title: Test
draft: true
source_lang: zh-TW
---
## Heading

Content here.

## Second

More content.
`), 0644)
}

func writeTargetFile(t *testing.T, path string) {
	t.Helper()
	os.WriteFile(path, []byte(`---
title: Test
draft: true
source_lang: zh-TW
target_lang: en
---
## Heading

Content translated here.

## Second

More content translated.
`), 0644)
}

func TestBatch_ProcessesMultipleFiles(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "one.md")
	src2 := filepath.Join(tmpDir, "src", "two.md")
	tgt1 := filepath.Join(tmpDir, "en", "one.md")
	tgt2 := filepath.Join(tmpDir, "en", "two.md")

	writeSourceFile(t, src1)
	writeSourceFile(t, src2)
	writeTargetFile(t, tgt1)
	writeTargetFile(t, tgt2)

	opts := core.BatchOptions{
		Provider: "ai-harness",
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 2 {
		t.Fatalf("expected 2 completed, got %d", len(report.Completed))
	}

	files, _ := content.Discover(cfg)
	completed := 0
	for _, f := range files {
		if f.Status == content.StatusExists {
			completed++
		}
	}
	if completed != 2 {
		t.Fatalf("expected 2 files with exists status, got %d", completed)
	}
}

func TestBatch_StopOnFail(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "aaa.md")
	src2 := filepath.Join(tmpDir, "src", "bbb.md")
	src3 := filepath.Join(tmpDir, "src", "ccc.md")
	tgt1 := filepath.Join(tmpDir, "en", "aaa.md")
	tgt2 := filepath.Join(tmpDir, "en", "bbb.md")

	writeSourceFile(t, src1)
	writeSourceFile(t, src2)
	writeSourceFile(t, src3)
	writeTargetFile(t, tgt1)

	os.WriteFile(tgt2, []byte(`---
title: Bad
draft: true
source_lang: zh-TW
target_lang: en
---
## Heading

CJK 中文 remain here.
`), 0644)

	opts := core.BatchOptions{
		Provider:   "ai-harness",
		StopOnFail: true,
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 1 {
		t.Fatalf("expected 1 completed (aaa), got %d", len(report.Completed))
	}
	if len(report.Failed) != 1 {
		t.Fatalf("expected 1 failed (bbb), got %d", len(report.Failed))
	}
	if len(report.Remaining) != 1 {
		t.Fatalf("expected 1 remaining (ccc), got %d", len(report.Remaining))
	}
}

func TestBatch_ContinueOnError(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "aaa.md")
	src2 := filepath.Join(tmpDir, "src", "bbb.md")
	src3 := filepath.Join(tmpDir, "src", "ccc.md")
	tgt1 := filepath.Join(tmpDir, "en", "aaa.md")
	tgt2 := filepath.Join(tmpDir, "en", "bbb.md")
	tgt3 := filepath.Join(tmpDir, "en", "ccc.md")

	writeSourceFile(t, src1)
	writeSourceFile(t, src2)
	writeSourceFile(t, src3)
	writeTargetFile(t, tgt1)
	writeTargetFile(t, tgt3)

	os.WriteFile(tgt2, []byte(`---
title: Bad
draft: true
source_lang: zh-TW
target_lang: en
---
## Heading

CJK 中文 remain.
`), 0644)

	opts := core.BatchOptions{
		Provider:        "ai-harness",
		ContinueOnError: true,
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 2 {
		t.Fatalf("expected 2 completed, got %d", len(report.Completed))
	}
	if len(report.Failed) != 1 {
		t.Fatalf("expected 1 failed, got %d", len(report.Failed))
	}
}

func TestBatch_SyncedCompletionsUpdateQueue(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "queue1.md")
	tgt1 := filepath.Join(tmpDir, "en", "queue1.md")

	writeSourceFile(t, src1)
	writeTargetFile(t, tgt1)

	statusBefore, _ := core.TranslationQueue(cfg, "")
	if statusBefore.Stale != 1 {
		t.Fatalf("expected 1 stale before batch, got %d", statusBefore.Stale)
	}

	opts := core.BatchOptions{
		Provider: "ai-harness",
	}
	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 1 {
		t.Fatalf("expected 1 completed, got %d", len(report.Completed))
	}

	statusAfter, _ := core.TranslationQueue(cfg, "")
	if statusAfter.Completed != 1 {
		t.Fatalf("expected 1 completed after batch, got %d", statusAfter.Completed)
	}
	if statusAfter.Stale != 0 {
		t.Fatalf("expected 0 stale after batch, got %d", statusAfter.Stale)
	}
}

func TestBatch_CliMcpAgreeOnFixture(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "fixture.md")
	tgt1 := filepath.Join(tmpDir, "en", "fixture.md")

	writeSourceFile(t, src1)
	writeTargetFile(t, tgt1)

	opts := core.BatchOptions{
		Provider: "ai-harness",
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 1 {
		t.Fatalf("expected 1 completed, got %d", len(report.Completed))
	}
	if report.Completed[0].Status != "completed" {
		t.Fatalf("expected status completed, got %s", report.Completed[0].Status)
	}

	status, _ := core.TranslationQueue(cfg, "")
	if status.Completed != 1 {
		t.Fatalf("queue should show 1 completed, got %d", status.Completed)
	}
}

func TestBatch_DryRun(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "dry.md")
	tgt1 := filepath.Join(tmpDir, "en", "dry.md")

	writeSourceFile(t, src1)

	opts := core.BatchOptions{
		Provider: "ai-harness",
		DryRun:   true,
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 0 {
		t.Fatalf("expected 0 completed in dry-run, got %d", len(report.Completed))
	}

	if _, err := os.Stat(tgt1); err == nil {
		t.Fatal("dry-run should not create target file")
	}
}

func TestBatch_PreservesUnknownFrontmatter(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "meta.md")
	tgt1 := filepath.Join(tmpDir, "en", "meta.md")

	os.WriteFile(src1, []byte(`---
title: Meta Test
draft: true
source_lang: zh-TW
date: 2025-01-15
slug: meta-test
tags:
  - tag1
  - tag2
categories:
  - dev
keywords: [key1, key2]
customField: customValue
---
## Heading

Content here.
`), 0644)

	os.WriteFile(tgt1, []byte(`---
title: Meta Test
draft: true
source_lang: zh-TW
target_lang: en
date: 2025-01-15
slug: meta-test
tags:
  - tag1
  - tag2
categories:
  - dev
keywords: [key1, key2]
customField: customValue
---
## Heading

Content translated here.
`), 0644)

	opts := core.BatchOptions{
		Provider: "ai-harness",
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 1 {
		t.Fatalf("expected 1 completed, got %d: %+v", len(report.Completed), report.Failed)
	}

	data, err := os.ReadFile(tgt1)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, field := range []string{"date:", "slug:", "tags:", "categories:", "keywords:", "customField:"} {
		if !strings.Contains(content, field) {
			t.Fatalf("expected frontmatter field %q preserved, got:\n%s", field, content)
		}
	}
}

type mockTranslator struct {
	translateFn func(text, sourceLang, targetLang string) (string, error)
}

func (m *mockTranslator) Translate(text, sourceLang, targetLang string) (string, error) {
	return m.translateFn(text, sourceLang, targetLang)
}

func TestBatch_InjectedTranslatorDeeplPath(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "deepl.md")
	tgt1 := filepath.Join(tmpDir, "en", "deepl.md")

	writeSourceFile(t, src1)

	mt := &mockTranslator{
		translateFn: func(text, sourceLang, targetLang string) (string, error) {
			return "## Heading\n\nTranslated by deepl.\n\n## Second\n\nMore translated content.", nil
		},
	}

	opts := core.BatchOptions{
		Provider:   "deepl",
		Translator: mt,
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 1 {
		t.Fatalf("expected 1 completed, got %d: %+v", len(report.Completed), report.Failed)
	}

	data, err := os.ReadFile(tgt1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Translated by deepl") {
		t.Fatalf("expected deepl translation in target, got:\n%s", string(data))
	}
}

func TestBatch_InjectedTranslatorGooglePath(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "google.md")
	tgt1 := filepath.Join(tmpDir, "en", "google.md")

	writeSourceFile(t, src1)

	mt := &mockTranslator{
		translateFn: func(text, sourceLang, targetLang string) (string, error) {
			return "## Heading\n\nTranslated by google.\n\n## Second\n\nMore translated content.", nil
		},
	}

	opts := core.BatchOptions{
		Provider:   "google",
		Translator: mt,
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 1 {
		t.Fatalf("expected 1 completed, got %d: %+v", len(report.Completed), report.Failed)
	}

	data, err := os.ReadFile(tgt1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Translated by google") {
		t.Fatalf("expected google translation in target, got:\n%s", string(data))
	}
}

func TestBatch_InjectedTranslatorAutoFallback(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "auto.md")
	tgt1 := filepath.Join(tmpDir, "en", "auto.md")

	writeSourceFile(t, src1)

	mt := &mockTranslator{
		translateFn: func(text, sourceLang, targetLang string) (string, error) {
			return "## Heading\n\nTranslated by auto.\n\n## Second\n\nMore translated content.", nil
		},
	}

	opts := core.BatchOptions{
		Provider:   "auto",
		Translator: mt,
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Completed) != 1 {
		t.Fatalf("expected 1 completed, got %d", len(report.Completed))
	}

	data, err := os.ReadFile(tgt1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Translated by auto") {
		t.Fatalf("expected auto translation in target, got:\n%s", string(data))
	}
}

func TestBatch_CJKGateCatchesFrontmatter(t *testing.T) {
	tmpDir, cfg := setupBatchTest(t)

	src1 := filepath.Join(tmpDir, "src", "cjk.md")
	tgt1 := filepath.Join(tmpDir, "en", "cjk.md")

	os.WriteFile(src1, []byte(`---
title: Test
draft: true
source_lang: zh-TW
---
## Heading

Content here.

## Second

More content.
`), 0644)

	os.WriteFile(tgt1, []byte(`---
title: Clean Title
draft: true
source_lang: zh-TW
target_lang: en
description: 中文描述
---
## Heading

Translated body is clean.

## Second

More translated content.
`), 0644)

	opts := core.BatchOptions{
		Provider: "ai-harness",
	}

	report, err := core.TranslateBatch(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Failed) != 1 {
		t.Fatalf("expected 1 failed (CJK in frontmatter), got %d completed, %d failed", len(report.Completed), len(report.Failed))
	}
	if !strings.Contains(report.Failed[0].Error, "CJK") {
		t.Fatalf("expected CJK error, got: %s", report.Failed[0].Error)
	}
}
