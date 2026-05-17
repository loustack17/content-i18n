package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/loustack17/content-i18n/internal/core"
)

func (s *Server) handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	report, err := core.Status(s.cfg)
	if err != nil {
		return errorResponse(err)
	}

	return jsonResponse(map[string]any{
		"project_type":              report.ProjectType,
		"source_language":           report.SourceLanguage,
		"source_path":               report.SourcePath,
		"target_languages":          report.TargetLanguages,
		"source_file_count":         report.SourceFileCount,
		"target_file_count":         report.TargetFileCount,
		"missing_translation_count": report.MissingCount,
	})
}

func (s *Server) handleValidateSite(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.cfg.Adapter.Name != core.AdapterHugo {
		return textResponse(fmt.Sprintf("validate-site only supports hugo adapter (got: %s)", s.cfg.Adapter.Name)), nil
	}

	warnings := core.ValidateSiteConfig(s.cfg)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	hugoRoot := s.cfg.ConfigDir
	result, err := core.ValidateSite(s.cfg, hugoRoot)
	if err != nil {
		return errorResponse(err)
	}

	if result.Passed {
		return textResponse("PASS"), nil
	}

	out, err := json.MarshalIndent(map[string]any{
		"hugo_output": result.HugoOutput,
		"violations":  result.Violations,
	}, "", "  ")
	if err != nil {
		return errorResponse(err)
	}
	return textResponse("FAIL\n" + string(out)), nil
}

func (s *Server) handlePrepareTranslation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, err := req.RequireString("source")
	if err != nil {
		return errorResponse(err)
	}
	lang, err := req.RequireString("language")
	if err != nil {
		return errorResponse(err)
	}

	result, err := core.TranslatePrepare(s.cfg, source, lang)
	if err != nil {
		return errorResponse(err)
	}

	return jsonResponse(map[string]any{
		"slug":        result.Slug,
		"source":      result.Source,
		"prompt":      result.Prompt,
		"glossary":    result.Glossary,
		"style":       result.Style,
		"context":     result.Context,
		"fingerprint": result.Fingerprint,
		"target_path": result.TargetPath,
	})
}

func (s *Server) handleReviewTranslation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, err := req.RequireString("source")
	if err != nil {
		return errorResponse(err)
	}
	target, err := req.RequireString("target")
	if err != nil {
		return errorResponse(err)
	}

	result, err := core.TranslateReview(s.cfg, source, target)
	if err != nil {
		return errorResponse(err)
	}

	return jsonResponse(map[string]any{
		"passed":        result.Passed,
		"ready_to_sync": result.ReadyToSync,
		"source_words":  result.SourceWords,
		"target_words":  result.TargetWords,
		"word_ratio":    result.WordRatio,
		"issues":        result.Issues,
	})
}

func (s *Server) handleSyncStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target, err := req.RequireString("target")
	if err != nil {
		return errorResponse(err)
	}
	source, err := req.RequireString("source")
	if err != nil {
		return errorResponse(err)
	}

	result, err := core.SyncStatus(s.cfg, target, source)
	if err != nil {
		return errorResponse(err)
	}

	return jsonResponse(map[string]any{
		"source":      result.SourcePath,
		"target":      result.TargetPath,
		"language":    result.Language,
		"source_hash": result.SourceHash,
		"status":      "synced",
	})
}

func (s *Server) handleTranslationQueue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	group := req.GetString("group", "")

	status, err := core.TranslationQueue(s.cfg, group)
	if err != nil {
		return errorResponse(err)
	}

	var next any
	if status.Next != nil {
		next = map[string]any{
			"source":   status.Next.SourcePath,
			"target":   status.Next.TargetPath,
			"language": status.Next.Language,
			"status":   string(status.Next.Status),
		}
	}

	return jsonResponse(map[string]any{
		"total":     status.Total,
		"completed": status.Completed,
		"stale":     status.Stale,
		"missing":   status.Missing,
		"next":      next,
	})
}

func (s *Server) handleTranslateBatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	group := req.GetString("group", "")
	provider := req.GetString("provider", "ai-harness")
	limit := int(req.GetFloat("limit", 0))
	stopOnFail := req.GetBool("stop_on_fail", false)
	continueOnError := req.GetBool("continue_on_error", false)
	dryRun := req.GetBool("dry_run", false)

	opts := core.BatchOptions{
		Group:           group,
		Provider:        provider,
		Limit:           limit,
		StopOnFail:      stopOnFail,
		ContinueOnError: continueOnError,
		DryRun:          dryRun,
	}

	report, err := core.TranslateBatch(s.cfg, opts)
	if err != nil {
		return errorResponse(err)
	}

	return jsonResponse(map[string]any{
		"total":     report.Total,
		"completed": report.Completed,
		"failed":    report.Failed,
		"remaining": report.Remaining,
	})
}
