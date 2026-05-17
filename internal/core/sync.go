package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
)

type SyncStatusResult struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Language   string `json:"language"`
	SourceHash string `json:"source_hash"`
}

func SyncStatus(cfg *config.Config, targetPath string, sourcePath string) (*SyncStatusResult, error) {
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("resolve source path: %w", err)
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, fmt.Errorf("resolve target path: %w", err)
	}

	if _, err := os.Stat(absSource); os.IsNotExist(err) {
		return nil, fmt.Errorf("source file does not exist: %s", absSource)
	}

	if _, err := os.Stat(absTarget); os.IsNotExist(err) {
		return nil, fmt.Errorf("target file does not exist: %s", absTarget)
	}

	lang, err := resolveLanguage(cfg, absTarget)
	if err != nil {
		return nil, err
	}

	expectedTarget, err := content.TargetPath(cfg, absSource, lang)
	if err != nil {
		return nil, fmt.Errorf("compute expected target path: %w", err)
	}

	if absTarget != expectedTarget {
		return nil, fmt.Errorf("target path mismatch: got %s, expected %s", absTarget, expectedTarget)
	}

	opts := &ValidateOptions{
		SourcePath: absSource,
		Config:     cfg,
	}
	result, err := ValidateContent(absTarget, opts)
	if err != nil {
		return nil, fmt.Errorf("validate target: %w", err)
	}
	if !result.Passed {
		return nil, fmt.Errorf("validation failed: %v", result.Violations)
	}

	srcHash, err := content.FileHash(absSource)
	if err != nil {
		return nil, fmt.Errorf("compute source hash: %w", err)
	}

	if err := content.WriteStatusEntry(cfg, absSource, lang, srcHash); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	return &SyncStatusResult{
		SourcePath: absSource,
		TargetPath: absTarget,
		Language:   lang,
		SourceHash: srcHash,
	}, nil
}

func resolveLanguage(cfg *config.Config, targetPath string) (string, error) {
	for lang, targetDir := range cfg.Paths.Targets {
		absDir, err := filepath.Abs(targetDir)
		if err != nil {
			continue
		}
		if hasPrefix(targetPath, absDir) {
			return lang, nil
		}
	}
	return "", fmt.Errorf("target file %s does not belong to any configured language directory", targetPath)
}

func hasPrefix(path, prefix string) bool {
	rel, err := filepath.Rel(prefix, path)
	if err != nil {
		return false
	}
	return rel != "" && rel != "." && !filepath.IsAbs(rel)
}
