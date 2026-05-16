package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/core"
)

func setupTestDir(t *testing.T) (string, *config.Config) {
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

func TestTranslatePrepare_ReturnsAllArtifacts(t *testing.T) {
	tmpDir, cfg := setupTestDir(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	sourceContent := `---
title: 測試
source_lang: zh-TW
draft: true
---
## 標題一

這是第一段。

## 標題二

- 項目一
- 項目二
`
	os.WriteFile(sourcePath, []byte(sourceContent), 0644)

	result, err := core.TranslatePrepare(cfg, sourcePath, "en")
	if err != nil {
		t.Fatal(err)
	}

	if result.Slug != "test" {
		t.Fatalf("expected slug 'test', got %q", result.Slug)
	}
	if !strings.Contains(result.Source, "標題一") {
		t.Fatal("expected source content")
	}
	if result.Prompt == "" {
		t.Fatal("expected prompt")
	}
	if result.Fingerprint.H2Count != 2 {
		t.Fatalf("expected 2 H2 headings, got %d", result.Fingerprint.H2Count)
	}
	if result.Context == "" {
		t.Fatal("expected context")
	}
}

func TestTranslateReview_PassesOnGoodTranslation(t *testing.T) {
	tmpDir, cfg := setupTestDir(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	targetPath := filepath.Join(tmpDir, "en", "test.md")

	os.WriteFile(sourcePath, []byte(`---
title: 測試
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading

Test paragraph.
`), 0644)

	os.WriteFile(targetPath, []byte(`---
title: Test
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading

Test paragraph.
`), 0644)

	result, err := core.TranslateReview(cfg, sourcePath, targetPath)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Passed {
		t.Fatalf("expected pass, got issues: %v", result.Issues)
	}
	if result.SourceWords == 0 {
		t.Fatal("expected non-zero source words")
	}
}

func TestTranslateReview_FailsOnStructureMismatch(t *testing.T) {
	tmpDir, cfg := setupTestDir(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	targetPath := filepath.Join(tmpDir, "en", "test.md")

	os.WriteFile(sourcePath, []byte(`---
title: 測試
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading One

## Heading Two

## Heading Three
`), 0644)

	os.WriteFile(targetPath, []byte(`---
title: Test
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading One

## Heading Two
`), 0644)

	result, err := core.TranslateReview(cfg, sourcePath, targetPath)
	if err != nil {
		t.Fatal(err)
	}

	if result.Passed {
		t.Fatal("expected fail for structure mismatch")
	}

	hasError := false
	for _, issue := range result.Issues {
		if issue.Severity == "error" && issue.Field == "structure" {
			hasError = true
		}
	}
	if !hasError {
		t.Fatalf("expected error-level structure issue, got: %v", result.Issues)
	}
}

func TestTranslateRepair_AcceptsValidContent(t *testing.T) {
	tmpDir, cfg := setupTestDir(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	os.WriteFile(sourcePath, []byte(`---
title: 測試
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading

Test paragraph.
`), 0644)

	slug := "repair-ok"
	workDir := filepath.Join("work", slug)
	os.MkdirAll(workDir, 0755)

	metaContent := `{"source_path":"` + sourcePath + `","target_language":"en","provider":"manual","structure_hash":"","fingerprint":{"heading_count":1,"h2_count":1,"h3_count":0,"h4_count":0,"ordered_list_count":0,"unordered_list_count":0,"table_count":0,"paragraph_count":1,"blockquote_count":0,"code_block_count":0}}`
	os.WriteFile(filepath.Join(workDir, "meta.json"), []byte(metaContent), 0644)

	repairContent := `---
title: Test
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading

Test paragraph.
`

	result, err := core.TranslateRepair(cfg, slug, repairContent)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Passed {
		t.Fatalf("expected repair to pass, got: %s", result.Message)
	}

	targetPath := filepath.Join(workDir, "target.md")
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatal("expected target.md to exist")
	}

	os.RemoveAll("work")
}

func TestTranslateRepair_RejectsInvalidContent(t *testing.T) {
	tmpDir, cfg := setupTestDir(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	os.WriteFile(sourcePath, []byte(`---
title: 測試
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading One

## Heading Two
`), 0644)

	slug := "repair-fail"
	workDir := filepath.Join("work", slug)
	os.MkdirAll(workDir, 0755)

	metaContent := `{"source_path":"` + sourcePath + `","target_language":"en","provider":"manual","structure_hash":"","fingerprint":{"heading_count":2,"h2_count":2,"h3_count":0,"h4_count":0,"ordered_list_count":0,"unordered_list_count":0,"table_count":0,"paragraph_count":0,"blockquote_count":0,"code_block_count":0}}`
	os.WriteFile(filepath.Join(workDir, "meta.json"), []byte(metaContent), 0644)

	repairContent := `---
title: Test
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading One
`

	result, err := core.TranslateRepair(cfg, slug, repairContent)
	if err != nil {
		t.Fatal(err)
	}

	if result.Passed {
		t.Fatalf("expected repair to fail, got: %s", result.Message)
	}
	if !strings.Contains(result.Message, "REPAIR FAILED") {
		t.Fatalf("expected REPAIR FAILED message, got: %s", result.Message)
	}

	os.RemoveAll("work")
}
