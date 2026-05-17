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
	if !strings.Contains(text, `"ready_to_sync": true`) {
		t.Fatalf("expected ready_to_sync=true, got: %s", text)
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
	if strings.Contains(text, `"ready_to_sync": true`) {
		t.Fatalf("expected ready_to_sync=false for structure error, got: %s", text)
	}
}

func TestEndToEnd_PrepareAndReview(t *testing.T) {
	srv, tmpDir := setupTestServer(t)

	sourcePath := filepath.Join(tmpDir, "src", "k8s-hpa-debug.md")
	sourceContent := `---
title: Kubernetes HPA 除錯記錄
source_lang: zh-TW
draft: true
---
## 問題描述

生產環境中 Horizontal Pod Autoscaler 沒有按預期擴展。

## 診斷步驟

使用以下命令檢查 HPA 狀態：

` + "```bash" + `
kubectl get hpa -n production
kubectl describe hpa my-app-hpa -n production
` + "```" + `

發現 metrics-server 沒有正確回報 CPU 使用率。

## 解決方案

1. 重新部署 metrics-server
2. 檢查 resource requests/limits
3. 驗證 HPA 的 targetCPUUtilizationPercentage

| 步驟 | 命令 | 預期結果 |
|------|------|----------|
| 1 | kubectl get pods | metrics-server running |
| 2 | kubectl top nodes | CPU metrics visible |
| 3 | kubectl get hpa | Scaling active |

> 注意：metrics-server 需要至少 2 分鐘才能收集足夠的指標。

## 後續改進

- 加入 Prometheus 監控
- 設定告警規則
- 文件化常見的 HPA 問題
`
	os.WriteFile(sourcePath, []byte(sourceContent), 0644)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"source":   sourcePath,
		"language": "en",
	}

	prepareResult, err := srv.handlePrepareTranslation(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	prepareText := prepareResult.Content[0].(mcp.TextContent).Text
	if !strings.Contains(prepareText, `"source"`) {
		t.Fatalf("prepare missing source content")
	}
	if !strings.Contains(prepareText, `"fingerprint"`) {
		t.Fatalf("prepare missing fingerprint")
	}
	if !strings.Contains(prepareText, `"prompt"`) {
		t.Fatalf("prepare missing prompt")
	}

	targetPath := filepath.Join(tmpDir, "en", "k8s-hpa-debug.md")
	goodTranslation := `---
title: Kubernetes HPA Debugging Log
source_lang: zh-TW
target_lang: en
draft: true
---
## Problem Description

The Horizontal Pod Autoscaler in production did not scale as expected.

## Diagnostic Steps

Use the following commands to check HPA status:

` + "```bash" + `
kubectl get hpa -n production
kubectl describe hpa my-app-hpa -n production
` + "```" + `

Found that metrics-server was not reporting CPU usage correctly.

## Solution

1. Redeploy metrics-server
2. Check resource requests/limits
3. Verify HPA targetCPUUtilizationPercentage

| Step | Command | Expected Result |
|------|---------|-----------------|
| 1 | kubectl get pods | metrics-server running |
| 2 | kubectl top nodes | CPU metrics visible |
| 3 | kubectl get hpa | Scaling active |

> Note: metrics-server needs at least 2 minutes to collect sufficient metrics.

## Follow-up Improvements

- Add Prometheus monitoring
- Configure alerting rules
- Document common HPA issues
`
	os.WriteFile(targetPath, []byte(goodTranslation), 0644)

	reviewReq := mcp.CallToolRequest{}
	reviewReq.Params.Arguments = map[string]any{
		"source": sourcePath,
		"target": targetPath,
	}

	reviewResult, err := srv.handleReviewTranslation(context.Background(), reviewReq)
	if err != nil {
		t.Fatal(err)
	}

	reviewText := reviewResult.Content[0].(mcp.TextContent).Text
	if !strings.Contains(reviewText, `"passed": true`) {
		t.Fatalf("expected good translation to pass review, got: %s", reviewText)
	}
	if !strings.Contains(reviewText, `"ready_to_sync": true`) {
		t.Fatalf("expected good translation ready_to_sync=true, got: %s", reviewText)
	}
}

func TestAllToolsRegistered(t *testing.T) {
	srv, _ := setupTestServer(t)

	specs := allToolSpecs(srv)
	for _, spec := range specs {
		name := spec.def.Name
		if name == "" {
			t.Fatal("tool definition has empty name")
		}
		if spec.handler == nil {
			t.Fatalf("tool %q has nil handler", name)
		}
	}

	expectedTools := []string{
		"content_i18n_status",
		"content_i18n_prepare_translation",
		"content_i18n_review_translation",
		"content_i18n_sync_status",
		"content_i18n_translation_queue",
		"content_i18n_translate_batch",
		"content_i18n_validate_site",
	}

	if len(specs) != len(expectedTools) {
		t.Fatalf("expected %d tool specs, got %d", len(expectedTools), len(specs))
	}

	for _, expected := range expectedTools {
		found := false
		for _, spec := range specs {
			if spec.def.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected tool %q not found", expected)
		}
	}
}

func TestAllResourcesRegistered(t *testing.T) {
	srv, _ := setupTestServer(t)

	specs := allResourceSpecs(srv)
	if len(specs) != 4 {
		t.Fatalf("expected 4 resource specs, got %d", len(specs))
	}

	expected := []string{
		"content-i18n://config",
		"content-i18n://glossary",
		"content-i18n://style-pack",
	}
	for _, uri := range expected {
		found := false
		for _, spec := range specs {
			switch d := spec.def.(type) {
			case mcp.Resource:
				if d.URI == uri {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("expected resource %q not found", uri)
		}
	}
}
