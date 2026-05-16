package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
)

type PrepareResult struct {
	Slug        string               `json:"slug"`
	Source      string               `json:"source"`
	Prompt      string               `json:"prompt"`
	Glossary    string               `json:"glossary"`
	Style       string               `json:"style"`
	Context     string               `json:"context"`
	Fingerprint StructureFingerprint `json:"fingerprint"`
	TargetPath  string               `json:"target_path"`
}

func TranslatePrepare(cfg *config.Config, sourceFile string, targetLang string) (*PrepareResult, error) {
	absSource, err := filepath.Abs(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("resolve source path: %w", err)
	}

	slug := SlugFromPath(absSource, cfg.Paths.Source)
	packet, err := GenerateWorkPacket(cfg, absSource, targetLang)
	if err != nil {
		return nil, fmt.Errorf("generate work packet: %w", err)
	}

	sourceData, err := os.ReadFile(absSource)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}

	var promptData, glossaryData, styleData, contextData string
	if data, err := os.ReadFile(packet.PromptPath); err == nil {
		promptData = string(data)
	}
	if data, err := os.ReadFile(packet.GlossaryPath); err == nil {
		glossaryData = string(data)
	}
	if data, err := os.ReadFile(packet.StylePath); err == nil {
		styleData = string(data)
	}
	if data, err := os.ReadFile(packet.ContextPath); err == nil {
		contextData = string(data)
	}

	metaData, err := os.ReadFile(packet.MetaPath)
	if err != nil {
		return nil, fmt.Errorf("read meta: %w", err)
	}
	var meta WorkMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, fmt.Errorf("parse meta: %w", err)
	}

	return &PrepareResult{
		Slug:        slug,
		Source:      string(sourceData),
		Prompt:      promptData,
		Glossary:    glossaryData,
		Style:       styleData,
		Context:     contextData,
		Fingerprint: meta.Fingerprint,
		TargetPath:  packet.TargetPath,
	}, nil
}

type ReviewIssue struct {
	Severity     string `json:"severity"`
	Field        string `json:"field"`
	Section      string `json:"section"`
	Message      string `json:"message"`
	SuggestedFix string `json:"suggested_fix"`
}

type ReviewResult struct {
	Passed      bool          `json:"passed"`
	SourceWords int           `json:"source_words"`
	TargetWords int           `json:"target_words"`
	WordRatio   string        `json:"word_ratio"`
	Issues      []ReviewIssue `json:"issues"`
}

func TranslateReview(cfg *config.Config, sourceFile string, targetFile string) (*ReviewResult, error) {
	absSource, err := filepath.Abs(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("resolve source path: %w", err)
	}
	absTarget, err := filepath.Abs(targetFile)
	if err != nil {
		return nil, fmt.Errorf("resolve target path: %w", err)
	}

	opts := &ValidateOptions{
		SourcePath: absSource,
		Config:     cfg,
	}
	result, err := ValidateContent(absTarget, opts)
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	sourceData, _ := os.ReadFile(absSource)
	targetData, _ := os.ReadFile(absTarget)
	srcWords := len(strings.Fields(string(sourceData)))
	tgtWords := len(strings.Fields(string(targetData)))

	var issues []ReviewIssue
	for _, v := range result.Violations {
		severity := "warning"
		if v.Field == "codeBlocks" || v.Field == "inlineCode" || v.Field == "urls" || v.Field == "structure" {
			severity = "error"
		}
		issues = append(issues, ReviewIssue{
			Severity:     severity,
			Field:        v.Field,
			Section:      v.Section,
			Message:      v.Message,
			SuggestedFix: v.SuggestedFix,
		})
	}

	ratio := "0%"
	if srcWords > 0 {
		ratio = fmt.Sprintf("%.0f%%", float64(tgtWords)/float64(srcWords)*100)
	}

	return &ReviewResult{
		Passed:      result.Passed,
		SourceWords: srcWords,
		TargetWords: tgtWords,
		WordRatio:   ratio,
		Issues:      issues,
	}, nil
}

type RepairResult struct {
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

func TranslateRepair(cfg *config.Config, slug string, content string) (*RepairResult, error) {
	workDir := filepath.Join("work", slug)
	metaPath := filepath.Join(workDir, "meta.json")

	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("read meta: %w", err)
	}
	var meta WorkMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, fmt.Errorf("parse meta: %w", err)
	}

	targetPath := filepath.Join(workDir, "target.md")

	tmpFile := targetPath + ".repair"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write temp: %w", err)
	}

	opts := &ValidateOptions{
		SourcePath: meta.SourcePath,
		Config:     cfg,
	}
	result, err := ValidateContent(tmpFile, opts)
	os.Remove(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	if !result.Passed {
		var lines []string
		for _, v := range result.Violations {
			lines = append(lines, fmt.Sprintf("[%s] %s: %s", v.Field, v.Section, v.Message))
		}
		return &RepairResult{
			Passed:  false,
			Message: "REPAIR FAILED\n" + strings.Join(lines, "\n"),
		}, nil
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("write target: %w", err)
	}

	return &RepairResult{
		Passed:  true,
		Message: fmt.Sprintf("REPAIR OK\nwrote %d bytes to %s", len(content), targetPath),
	}, nil
}
