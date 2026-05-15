package core

import (
	"fmt"
	"path/filepath"

	"github.com/loustack17/content-i18n/internal/validator"
)

func ValidateContent(targetFile string) error {
	abs, err := filepath.Abs(targetFile)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	violations, err := validator.Validate(abs, "")
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Printf("[%s] %s\n", v.Field, v.Message)
		}
		return fmt.Errorf("validation failed with %d issue(s)", len(violations))
	}

	fmt.Printf("validation passed: %s\n", abs)
	return nil
}

func ValidateSite(cfgPath string) error {
	return fmt.Errorf("validate-site: not yet implemented")
}
