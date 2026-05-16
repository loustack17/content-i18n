package core_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/core"
	"github.com/stretchr/testify/require"
)

func fileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func writeStatus(t *testing.T, cfg *config.Config, entries map[string]string) {
	t.Helper()
	statusDir := filepath.Dir(cfg.StatusFilePath())
	require.NoError(t, os.MkdirAll(statusDir, 0755))
	store := map[string]any{"entries": entries}
	data, err := json.MarshalIndent(store, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfg.StatusFilePath(), data, 0644))
}

func setupQueueTest(t *testing.T) (string, *config.Config) {
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

func TestQueue_NextSkipsCompletedFiles(t *testing.T) {
	tmpDir, cfg := setupQueueTest(t)

	os.WriteFile(filepath.Join(tmpDir, "src", "done.md"), []byte(`---
title: Done
draft: true
---
## Heading

Content.
`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "en", "done.md"), []byte(`---
title: Done
draft: true
---
## Heading

Content.
`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "src", "pending.md"), []byte(`---
title: Pending
draft: true
---
## Heading

Content.
`), 0644)

	hash := fileHash(filepath.Join(tmpDir, "src", "done.md"))
	writeStatus(t, cfg, map[string]string{"done.md:en": hash})

	entry, err := core.NextTranslation(cfg, "")
	if err != nil {
		t.Fatal(err)
	}

	if entry == nil {
		t.Fatal("expected next entry, got nil")
	}
	if filepath.Base(entry.SourcePath) != "pending.md" {
		t.Fatalf("expected pending.md, got %s", entry.SourcePath)
	}
}

func TestQueue_StaleFilesReappear(t *testing.T) {
	tmpDir, cfg := setupQueueTest(t)

	sourcePath := filepath.Join(tmpDir, "src", "stale.md")
	os.WriteFile(sourcePath, []byte(`---
title: Stale
draft: true
---
## Heading

Content.
`), 0644)

	targetPath := filepath.Join(tmpDir, "en", "stale.md")
	os.WriteFile(targetPath, []byte(`---
title: Stale
draft: true
---
## Heading

Old content.
`), 0644)

	writeStatus(t, cfg, map[string]string{"stale.md:en": "old_hash_value"})

	entry, err := core.NextTranslation(cfg, "")
	if err != nil {
		t.Fatal(err)
	}

	if entry == nil {
		t.Fatal("expected stale file to reappear in queue")
	}
	if filepath.Base(entry.SourcePath) != "stale.md" {
		t.Fatalf("expected stale.md, got %s", entry.SourcePath)
	}
}

func TestQueue_BatchStatusCountsCorrect(t *testing.T) {
	tmpDir, cfg := setupQueueTest(t)

	os.WriteFile(filepath.Join(tmpDir, "src", "done.md"), []byte(`---
title: Done
draft: true
---
## Heading

Content.
`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "en", "done.md"), []byte(`---
title: Done
draft: true
---
## Heading

Content.
`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "src", "missing.md"), []byte(`---
title: Missing
draft: true
---
## Heading

Content.
`), 0644)

	hash := fileHash(filepath.Join(tmpDir, "src", "done.md"))
	writeStatus(t, cfg, map[string]string{"done.md:en": hash})

	status, err := core.TranslationQueue(cfg, "")
	if err != nil {
		t.Fatal(err)
	}

	if status.Total != 2 {
		t.Fatalf("expected total 2, got %d", status.Total)
	}
	if status.Completed != 1 {
		t.Fatalf("expected completed 1, got %d", status.Completed)
	}
	if status.Missing != 1 {
		t.Fatalf("expected missing 1, got %d", status.Missing)
	}
}

func TestQueue_MCPCliAgreeOnNextFile(t *testing.T) {
	tmpDir, cfg := setupQueueTest(t)

	os.WriteFile(filepath.Join(tmpDir, "src", "aaa.md"), []byte(`---
title: AAA
draft: true
---
## Heading

Content.
`), 0644)

	os.WriteFile(filepath.Join(tmpDir, "src", "bbb.md"), []byte(`---
title: BBB
draft: true
---
## Heading

Content.
`), 0644)

	entry, err := core.NextTranslation(cfg, "")
	if err != nil {
		t.Fatal(err)
	}

	if entry == nil {
		t.Fatal("expected next entry")
	}

	status, err := core.TranslationQueue(cfg, "")
	if err != nil {
		t.Fatal(err)
	}

	if status.Next == nil {
		t.Fatal("expected next in status")
	}

	if entry.SourcePath != status.Next.SourcePath {
		t.Fatalf("CLI next %s != MCP next %s", entry.SourcePath, status.Next.SourcePath)
	}
}
