package validator

import (
	"fmt"
	"os"
	"strings"

	"github.com/loustack17/content-i18n/internal/frontmatter"
)

type Violation struct {
	Field   string
	Message string
}

func Validate(targetPath string, sourcePath string) ([]Violation, error) {
	var violations []Violation

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("read target: %w", err)
	}

	targetContent := string(targetData)

	if !strings.HasPrefix(strings.TrimSpace(targetContent), "---") {
		violations = append(violations, Violation{Field: "frontmatter", Message: "target missing frontmatter"})
	}

	targetDoc := frontmatter.Split(targetContent)

	if targetDoc.Metadata.Title == "" {
		violations = append(violations, Violation{Field: "title", Message: "target title missing"})
	}

	if sourcePath == "" {
		return violations, nil
	}

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}

	sourceContent := string(sourceData)
	sourceDoc := frontmatter.Split(sourceContent)

	if sourceDoc.Metadata.TranslationKey != "" && targetDoc.Metadata.TranslationKey != sourceDoc.Metadata.TranslationKey {
		violations = append(violations, Violation{Field: "translationKey", Message: "translationKey mismatch"})
	}

	sourceCodeBlocks := frontmatter.ExtractCodeBlocks(sourceContent)
	targetCodeBlocks := frontmatter.ExtractCodeBlocks(targetContent)
	if len(targetCodeBlocks) < len(sourceCodeBlocks) {
		violations = append(violations, Violation{Field: "codeBlocks", Message: fmt.Sprintf("target has %d code blocks, source has %d", len(targetCodeBlocks), len(sourceCodeBlocks))})
	}

	sourceInline := frontmatter.ExtractInlineCode(sourceContent)
	targetInline := frontmatter.ExtractInlineCode(targetContent)
	if len(targetInline) < len(sourceInline) {
		violations = append(violations, Violation{Field: "inlineCode", Message: fmt.Sprintf("target has %d inline code, source has %d", len(targetInline), len(sourceInline))})
	}

	return violations, nil
}
