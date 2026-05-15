package core

import (
	"fmt"
	"path/filepath"

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

	glossaryPath := opts.GlossaryPath
	if glossaryPath == "" && opts.Config != nil && opts.Config.Style.Glossary != "" {
		glossaryPath = opts.Config.Style.Glossary
		if !filepath.IsAbs(glossaryPath) {
			glossaryPath = filepath.Join(opts.Config.ConfigDir, glossaryPath)
		}
	}

	vOpts := &validator.ValidateOptions{
		GlossaryPath: glossaryPath,
		BannedWords:  opts.Config.Style.BannedWords,
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

func ValidateSite(cfgPath string) error {
	return fmt.Errorf("validate-site: not yet implemented")
}
