package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/loustack17/content-i18n/internal/config"
)

func setupTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	tmpDir := t.TempDir()

	cfgPath := filepath.Join(tmpDir, "content-i18n.yaml")
	cfgData := `project:
  type: generic-markdown
  source_language: zh-TW
  target_languages:
    - en
paths:
  source: ` + tmpDir + `/src
  targets:
    en: ` + tmpDir + `/en
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
	if err := os.WriteFile(cfgPath, []byte(cfgData), 0644); err != nil {
		t.Fatal(err)
	}

	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "en"), 0755)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	return NewServer(cfg, cfgPath), tmpDir
}

func TestPrepareTranslation_ReturnsSourceAndFingerprint(t *testing.T) {
	srv, tmpDir := setupTestServer(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	sourceContent := `---
title: 測試
draft: true
---
## 標題一

這是第一段。

## 標題二

- 項目一
- 項目二

`
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0644); err != nil {
		t.Fatal(err)
	}

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"source":   sourcePath,
		"language": "en",
	}

	result, err := srv.handlePrepareTranslation(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text

	if !strings.Contains(text, `"source"`) {
		t.Fatalf("expected source in result, got: %s", text)
	}
	if !strings.Contains(text, `"fingerprint"`) {
		t.Fatalf("expected fingerprint in result, got: %s", text)
	}
	if !strings.Contains(text, `"prompt"`) {
		t.Fatalf("expected prompt in result, got: %s", text)
	}
	if !strings.Contains(text, `"slug"`) {
		t.Fatalf("expected slug in result, got: %s", text)
	}
}

func TestReviewTranslation_PassesOnMatchingContent(t *testing.T) {
	srv, tmpDir := setupTestServer(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	targetPath := filepath.Join(tmpDir, "en", "test.md")

	sourceContent := `---
title: 測試
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading

Test paragraph.
`
	targetContent := `---
title: Test
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading

Test paragraph.
`
	os.WriteFile(sourcePath, []byte(sourceContent), 0644)
	os.WriteFile(targetPath, []byte(targetContent), 0644)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"source": sourcePath,
		"target": targetPath,
	}

	result, err := srv.handleReviewTranslation(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"passed": true`) {
		t.Fatalf("expected passed=true, got: %s", text)
	}
}

func TestReviewTranslation_FailsOnStructureMismatch(t *testing.T) {
	srv, tmpDir := setupTestServer(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	targetPath := filepath.Join(tmpDir, "en", "test.md")

	sourceContent := `---
title: 測試
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading One

## Heading Two

## Heading Three
`
	targetContent := `---
title: Test
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading One

## Heading Two
`
	os.WriteFile(sourcePath, []byte(sourceContent), 0644)
	os.WriteFile(targetPath, []byte(targetContent), 0644)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"source": sourcePath,
		"target": targetPath,
	}

	result, err := srv.handleReviewTranslation(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if strings.Contains(text, `"passed": true`) {
		t.Fatalf("expected passed=false for structure mismatch, got: %s", text)
	}
	if !strings.Contains(text, `"severity": "error"`) {
		t.Fatalf("expected error severity, got: %s", text)
	}
}

func TestRepairTranslation_AcceptsValidContent(t *testing.T) {
	srv, tmpDir := setupTestServer(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	sourceContent := `---
title: 測試
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading

Test paragraph.
`
	os.WriteFile(sourcePath, []byte(sourceContent), 0644)

	slug := "test-repair-ok"
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
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":    slug,
		"content": repairContent,
	}

	result, err := srv.handleRepairTranslation(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "REPAIR OK") {
		t.Fatalf("expected REPAIR OK, got: %s", text)
	}

	targetPath := filepath.Join(workDir, "target.md")
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("expected target.md to exist")
	}

	os.RemoveAll("work")
}

func TestRepairTranslation_RejectsInvalidContent(t *testing.T) {
	srv, tmpDir := setupTestServer(t)

	sourcePath := filepath.Join(tmpDir, "src", "test.md")
	sourceContent := `---
title: 測試
source_lang: zh-TW
target_lang: en
draft: true
---
## Heading One

## Heading Two
`
	os.WriteFile(sourcePath, []byte(sourceContent), 0644)

	slug := "test-repair-fail"
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
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"slug":    slug,
		"content": repairContent,
	}

	result, err := srv.handleRepairTranslation(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "REPAIR FAILED") {
		t.Fatalf("expected REPAIR FAILED, got: %s", text)
	}

	os.RemoveAll("work")
}
