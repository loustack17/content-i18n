package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/validator"
)

type ValidateOptions struct {
	SourcePath   string
	GlossaryPath string
	Config       *config.Config
}

type ValidateResult struct {
	Passed     bool
	Violations []validator.Violation
}

func ValidateContent(targetFile string, opts *ValidateOptions) (*ValidateResult, error) {
	abs, err := filepath.Abs(targetFile)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	if opts == nil {
		opts = &ValidateOptions{}
	}

	glossaryPath := opts.GlossaryPath
	if glossaryPath == "" && opts.Config != nil && opts.Config.Style.Glossary != "" {
		glossaryPath = opts.Config.Style.Glossary
		if !filepath.IsAbs(glossaryPath) {
			glossaryPath = filepath.Join(opts.Config.ConfigDir, glossaryPath)
		}
	}

	var bannedWords []string
	if opts.Config != nil {
		bannedWords = opts.Config.Style.BannedWords
	}

	vOpts := &validator.ValidateOptions{
		GlossaryPath: glossaryPath,
		BannedWords:  bannedWords,
	}

	violations, err := validator.Validate(abs, opts.SourcePath, vOpts)
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return &ValidateResult{
		Passed:     len(violations) == 0,
		Violations: violations,
	}, nil
}

type SiteValidateResult struct {
	Passed     bool
	HugoOutput string
	Violations []string
}

func ValidateSite(cfg *config.Config, hugoRoot string) (*SiteValidateResult, error) {
	if _, err := exec.LookPath("hugo"); err != nil {
		return nil, fmt.Errorf("hugo not found in PATH: %w", err)
	}

	cmd := exec.Command("hugo", "--minify", "--source", hugoRoot)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("hugo build failed: %s\n%s", err, string(output))
	}

	publicDir := filepath.Join(hugoRoot, "public")
	return validateSitePaths(cfg, publicDir)
}

func validateSitePaths(cfg *config.Config, publicDir string) (*SiteValidateResult, error) {
	result := &SiteValidateResult{
		HugoOutput: publicDir,
	}

	if _, err := os.Stat(publicDir); err != nil {
		return nil, fmt.Errorf("public directory not found: %s", publicDir)
	}

	canonical := cfg.URLPolicy.Canonical
	for lang, expectedPath := range canonical {
		fullPath := filepath.Join(publicDir, expectedPath)
		if _, err := os.Stat(fullPath); err != nil {
			result.Violations = append(result.Violations, fmt.Sprintf("expected canonical path for %s not found: %s", lang, expectedPath))
		}
	}

	for _, lang := range cfg.Project.TargetLanguages {
		if _, isCanonical := canonical[lang]; !isCanonical {
			unexpectedPath := filepath.Join(lang, "index.html")
			fullPath := filepath.Join(publicDir, lang, "index.html")
			if _, err := os.Stat(fullPath); err == nil {
				result.Violations = append(result.Violations, fmt.Sprintf("unexpected canonical path for non-canonical language %s: %s", lang, unexpectedPath))
			}
		}
	}

	if cfg.URLPolicy.RuntimeTranslation.Enabled {
		rtRoute := cfg.URLPolicy.RuntimeTranslation.Route
		if rtRoute == "" {
			result.Violations = append(result.Violations, "runtime translation enabled but route not configured")
		}
	}

	result.Passed = len(result.Violations) == 0
	return result, nil
}

func ValidateSiteConfig(cfg *config.Config) []string {
	var warnings []string

	if cfg.Adapter.Name == "hugo" {
		if cfg.URLPolicy.Canonical == nil || len(cfg.URLPolicy.Canonical) == 0 {
			warnings = append(warnings, "hugo adapter requires url_policy.canonical to be set")
		}
	}

	if cfg.URLPolicy.RuntimeTranslation.Enabled {
		if strings.TrimSpace(cfg.URLPolicy.RuntimeTranslation.Route) == "" {
			warnings = append(warnings, "runtime_translation enabled but route is empty")
		}
		if len(cfg.URLPolicy.RuntimeTranslation.CanonicalLanguages) == 0 {
			warnings = append(warnings, "runtime_translation enabled but canonical_languages is empty")
		}
	}

	return warnings
}
