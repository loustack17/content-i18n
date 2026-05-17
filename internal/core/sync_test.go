package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
	"github.com/loustack17/content-i18n/internal/core"
)

func setupSyncTest(t *testing.T) (string, *config.Config) {
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

func TestSyncStatus_MarksValidTargetComplete(t *testing.T) {
	tmpDir, cfg := setupSyncTest(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	targetPath := filepath.Join(tmpDir, "en", "test.md")

	os.WriteFile(sourcePath, []byte(`---
title: Test
draft: true
source_lang: zh-TW
---
## Heading

Content here.
`), 0644)

	os.WriteFile(targetPath, []byte(`---
title: Test
draft: true
source_lang: zh-TW
target_lang: en
---
## Heading

Content here translated.
`), 0644)

	result, err := core.SyncStatus(cfg, targetPath, sourcePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Language != "en" {
		t.Fatalf("expected language en, got %s", result.Language)
	}
	if result.SourceHash == "" {
		t.Fatal("expected non-empty source hash")
	}

	files, err := content.Discover(cfg)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range files {
		if f.SourcePath == sourcePath && f.Language == "en" {
			if f.Status != content.StatusExists {
				t.Fatalf("expected status exists, got %s", f.Status)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("source file not found in discovery")
	}
}

func TestSyncStatus_RejectsMissingTarget(t *testing.T) {
	tmpDir, cfg := setupSyncTest(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	targetPath := filepath.Join(tmpDir, "en", "test.md")

	os.WriteFile(sourcePath, []byte("source"), 0644)

	_, err := core.SyncStatus(cfg, targetPath, sourcePath)
	if err == nil {
		t.Fatal("expected error for missing target, got nil")
	}
}

func TestSyncStatus_RejectsWrongSourceTargetPair(t *testing.T) {
	tmpDir, cfg := setupSyncTest(t)

	sourcePath := filepath.Join(tmpDir, "src", "real.md")
	otherSource := filepath.Join(tmpDir, "src", "other.md")
	targetPath := filepath.Join(tmpDir, "en", "real.md")

	os.WriteFile(sourcePath, []byte("real source"), 0644)
	os.WriteFile(otherSource, []byte("other source"), 0644)
	os.WriteFile(targetPath, []byte("translated content"), 0644)

	_, err := core.SyncStatus(cfg, targetPath, otherSource)
	if err == nil {
		t.Fatal("expected error for mismatched source/target pair, got nil")
	}
}

func TestSyncStatus_StaleSourceReappearsAfterChange(t *testing.T) {
	tmpDir, cfg := setupSyncTest(t)

	sourcePath := filepath.Join(tmpDir, "src", "change.md")
	targetPath := filepath.Join(tmpDir, "en", "change.md")

	os.WriteFile(sourcePath, []byte(`---
title: Change
draft: true
source_lang: zh-TW
---
## Heading

Original content.
`), 0644)

	os.WriteFile(targetPath, []byte(`---
title: Change
draft: true
source_lang: zh-TW
target_lang: en
---
## Heading

Translated original.
`), 0644)

	_, err := core.SyncStatus(cfg, targetPath, sourcePath)
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	entry, err := core.NextTranslation(cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if entry != nil {
		t.Fatal("expected queue empty after sync, but got entry")
	}

	os.WriteFile(sourcePath, []byte(`---
title: Change
draft: true
---
## Heading

Modified content after translation.
`), 0644)

	entry, err = core.NextTranslation(cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if entry == nil {
		t.Fatal("expected stale file to reappear after source change")
	}
	if filepath.Base(entry.SourcePath) != "change.md" {
		t.Fatalf("expected change.md, got %s", entry.SourcePath)
	}
}

func TestSyncStatus_QueueReflectsSyncedCompletion(t *testing.T) {
	tmpDir, cfg := setupSyncTest(t)

	sourcePath := filepath.Join(tmpDir, "src", "queue.md")
	targetPath := filepath.Join(tmpDir, "en", "queue.md")

	os.WriteFile(sourcePath, []byte(`---
title: Queue
draft: true
source_lang: zh-TW
---
## Heading

Queue test content.
`), 0644)

	os.WriteFile(targetPath, []byte(`---
title: Queue
draft: true
source_lang: zh-TW
target_lang: en
---
## Heading

Queue test content translated.
`), 0644)

	statusBefore, err := core.TranslationQueue(cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if statusBefore.Stale != 1 {
		t.Fatalf("expected 1 stale before sync, got %d", statusBefore.Stale)
	}

	_, err = core.SyncStatus(cfg, targetPath, sourcePath)
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	statusAfter, err := core.TranslationQueue(cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if statusAfter.Completed != 1 {
		t.Fatalf("expected 1 completed after sync, got %d", statusAfter.Completed)
	}
	if statusAfter.Stale != 0 {
		t.Fatalf("expected 0 stale after sync, got %d", statusAfter.Stale)
	}
	if statusAfter.Next != nil {
		t.Fatal("expected no next entry after sync, but got one")
	}
}

func TestSyncStatus_RejectsTargetFailingConfigValidation(t *testing.T) {
	tmpDir, cfg := setupSyncTest(t)

	glossaryPath := filepath.Join(tmpDir, "glossary.yaml")
	os.WriteFile(glossaryPath, []byte(`terms:
  - source: "Kubernetes"
    target: "Kubernetes"
`), 0644)

	cfg.Style.Glossary = glossaryPath

	sourcePath := filepath.Join(tmpDir, "src", "gloss.md")
	targetPath := filepath.Join(tmpDir, "en", "gloss.md")

	os.WriteFile(sourcePath, []byte(`---
title: Gloss
draft: true
source_lang: zh-TW
---
## Heading

Use Kubernetes for orchestration.
`), 0644)

	os.WriteFile(targetPath, []byte(`---
title: Gloss
draft: true
source_lang: zh-TW
target_lang: en
---
## Heading

Use K8s for orchestration.
`), 0644)

	_, err := core.SyncStatus(cfg, targetPath, sourcePath)
	if err == nil {
		t.Fatal("expected error for glossary violation, got nil")
	}
}
