package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/frontmatter"
	"github.com/loustack17/content-i18n/internal/providers/deepl"
	"github.com/loustack17/content-i18n/internal/providers/google"
)

type BatchFileResult struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Language   string `json:"language"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

type BatchReport struct {
	Completed []BatchFileResult `json:"completed"`
	Failed    []BatchFileResult `json:"failed"`
	Remaining []BatchFileResult `json:"remaining"`
	Total     int               `json:"total"`
}

type BatchOptions struct {
	Group           string
	Provider        string
	Limit           int
	StopOnFail      bool
	ContinueOnError bool
	DryRun          bool
	Translator      Translator
}

type Translator interface {
	Translate(text string, sourceLang string, targetLang string) (string, error)
}

func TranslateBatch(cfg *config.Config, opts BatchOptions) (*BatchReport, error) {
	queueStatus, err := TranslationQueue(cfg, opts.Group)
	if err != nil {
		return nil, err
	}

	entries := queueStatus.Queue
	if opts.Limit > 0 && len(entries) > opts.Limit {
		entries = entries[:opts.Limit]
	}

	if len(entries) == 0 {
		return &BatchReport{Total: queueStatus.Total}, nil
	}

	var translator Translator
	if opts.Translator != nil {
		translator = opts.Translator
	} else if opts.Provider != "ai-harness" {
		translator, err = createTranslator(opts.Provider)
		if err != nil {
			return nil, err
		}
	}

	report := &BatchReport{Total: queueStatus.Total}

	for _, entry := range entries {
		result, err := processFile(cfg, entry, translator, opts)
		if err != nil {
			return nil, err
		}

		switch result.Status {
		case "completed":
			report.Completed = append(report.Completed, result)
		case "failed":
			report.Failed = append(report.Failed, result)
			if opts.StopOnFail {
				report.Remaining = appendRemaining(entries, len(report.Completed)+len(report.Failed))
				return report, nil
			}
		case "pending":
			report.Remaining = append(report.Remaining, result)
		}
	}

	return report, nil
}

func createTranslator(provider string) (Translator, error) {
	switch provider {
	case "deepl":
		return deepl.New()
	case "google":
		return google.New()
	case "auto":
		if p, err := deepl.New(); err == nil {
			return p, nil
		}
		return google.New()
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func processFile(cfg *config.Config, entry QueueEntry, translator Translator, opts BatchOptions) (BatchFileResult, error) {
	result := BatchFileResult{
		SourcePath: entry.SourcePath,
		TargetPath: entry.TargetPath,
		Language:   entry.Language,
	}

	if opts.DryRun {
		result.Status = "dry-run"
		return result, nil
	}

	if _, err := GenerateWorkPacket(cfg, entry.SourcePath, entry.Language); err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("prepare: %v", err)
		return result, nil
	}

	if translator != nil {
		sourceData, err := os.ReadFile(entry.SourcePath)
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("read source: %v", err)
			return result, nil
		}

		sourceDoc, err := frontmatter.Split(string(sourceData))
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("parse source frontmatter: %v", err)
			return result, nil
		}
		translatedBody, err := translator.Translate(sourceDoc.Body, cfg.Project.SourceLanguage, entry.Language)
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("translate: %v", err)
			return result, nil
		}

		targetDoc := sourceDoc
		targetDoc.Body = translatedBody
		targetDoc.RawMeta["target_lang"] = entry.Language
		targetContent, err := frontmatter.InjectProviderMeta(targetDoc, frontmatter.ProviderMeta{
			Provider: opts.Provider,
			Quality:  "machine_draft",
			Reviewed: false,
			Draft:    true,
		})
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("inject provider metadata: %v", err)
			return result, nil
		}

		if err := os.MkdirAll(filepath.Dir(entry.TargetPath), 0755); err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("create target dir: %v", err)
			return result, nil
		}
		if err := os.WriteFile(entry.TargetPath, []byte(targetContent), 0644); err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("write target: %v", err)
			return result, nil
		}
	} else {
		targetData, err := os.ReadFile(entry.TargetPath)
		if err != nil || len(strings.TrimSpace(string(targetData))) == 0 {
			result.Status = "pending"
			result.Error = "target not filled by AI agent"
			return result, nil
		}
	}

	absTarget := entry.TargetPath
	if !filepath.IsAbs(absTarget) {
		absTarget, _ = filepath.Abs(absTarget)
	}

	reviewResult, err := TranslateReview(cfg, entry.SourcePath, absTarget)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("review: %v", err)
		return result, nil
	}

	if !reviewResult.Passed && !opts.ContinueOnError {
		result.Status = "failed"
		result.Error = fmt.Sprintf("review failed: %d issues", len(reviewResult.Issues))
		return result, nil
	}

	if !reviewResult.Passed && opts.ContinueOnError {
		targetData, _ := os.ReadFile(absTarget)
		repairResult, err := TranslateRepair(cfg, SlugFromPath(entry.SourcePath, cfg.Paths.Source), string(targetData))
		if err != nil || !repairResult.Passed {
			result.Status = "failed"
			result.Error = fmt.Sprintf("repair failed: %v", err)
			return result, nil
		}
	}

	hasCJK, err := hasCJKInTarget(absTarget, entry.Language)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("CJK check: %v", err)
		return result, nil
	}
	if hasCJK {
		result.Status = "failed"
		result.Error = "CJK characters remain in target"
		return result, nil
	}

	_, err = SyncStatus(cfg, absTarget, entry.SourcePath)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("sync-status: %v", err)
		return result, nil
	}

	result.Status = "completed"
	return result, nil
}

func hasCJKInTarget(targetPath string, targetLang string) (bool, error) {
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return false, fmt.Errorf("read target for CJK check: %w", err)
	}
	doc, err := frontmatter.Split(string(data))
	if err != nil {
		return false, fmt.Errorf("parse target frontmatter for CJK check: %w", err)
	}
	for _, r := range doc.Body {
		if isCJK(r) && !unicode.IsSpace(r) {
			return true, nil
		}
	}
	for _, v := range doc.RawMeta {
		if s, ok := v.(string); ok {
			for _, r := range s {
				if isCJK(r) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func isCJK(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF ||
		r >= 0x3400 && r <= 0x4DBF ||
		r >= 0x20000 && r <= 0x2A6DF ||
		r >= 0xF900 && r <= 0xFAFF ||
		r >= 0x2F800 && r <= 0x2FA1F ||
		r >= 0x3040 && r <= 0x30FF ||
		r >= 0xAC00 && r <= 0xD7AF
}

func appendRemaining(entries []QueueEntry, processed int) []BatchFileResult {
	var remaining []BatchFileResult
	for i := processed; i < len(entries); i++ {
		remaining = append(remaining, BatchFileResult{
			SourcePath: entries[i].SourcePath,
			TargetPath: entries[i].TargetPath,
			Language:   entries[i].Language,
			Status:     "remaining",
		})
	}
	return remaining
}
