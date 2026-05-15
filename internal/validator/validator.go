package validator

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

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

	if targetDoc.Metadata.Draft && !targetDoc.Metadata.Reviewed {
	} else if !targetDoc.Metadata.Draft && !targetDoc.Metadata.Reviewed {
		violations = append(violations, Violation{Field: "draft", Message: "draft should be true until reviewed"})
	}

	cjkInTitle := countCJK(targetDoc.Metadata.Title)
	if cjkInTitle > 0 {
		violations = append(violations, Violation{Field: "title", Message: fmt.Sprintf("title contains %d CJK character(s)", cjkInTitle)})
	}

	cjkRatio := cjkRatioInBody(targetDoc.Body)
	if cjkRatio > 0.05 {
		violations = append(violations, Violation{Field: "language", Message: fmt.Sprintf("CJK ratio %.1f%% exceeds 5%% threshold", cjkRatio*100)})
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
	if len(targetCodeBlocks) != len(sourceCodeBlocks) {
		violations = append(violations, Violation{Field: "codeBlocks", Message: fmt.Sprintf("target has %d code blocks, source has %d", len(targetCodeBlocks), len(sourceCodeBlocks))})
	} else {
		for i := range sourceCodeBlocks {
			if strings.TrimSpace(sourceCodeBlocks[i]) != strings.TrimSpace(targetCodeBlocks[i]) {
				violations = append(violations, Violation{Field: "codeBlocks", Message: fmt.Sprintf("code block %d content differs from source", i+1)})
				break
			}
		}
	}

	sourceInline := frontmatter.ExtractInlineCode(sourceContent)
	targetInline := frontmatter.ExtractInlineCode(targetContent)
	if len(targetInline) < len(sourceInline) {
		violations = append(violations, Violation{Field: "inlineCode", Message: fmt.Sprintf("target has %d inline code, source has %d", len(targetInline), len(sourceInline))})
	}

	if len(sourceInline) == len(targetInline) {
		for i := range sourceInline {
			if strings.TrimSpace(sourceInline[i]) != strings.TrimSpace(targetInline[i]) {
				violations = append(violations, Violation{Field: "inlineCode", Message: fmt.Sprintf("inline code %q changed to %q", sourceInline[i], targetInline[i])})
				break
			}
		}
	}

	return violations, nil
}

func countCJK(s string) int {
	count := 0
	for _, r := range s {
		if isCJK(r) {
			count++
		}
	}
	return count
}

func cjkRatioInBody(body string) float64 {
	total := 0
	cjk := 0
	for _, r := range body {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || isCJK(r) {
			total++
			if isCJK(r) {
				cjk++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(cjk) / float64(total)
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

var urlPattern = regexp.MustCompile(`https?://[^\s\)\]"<>]+`)

func ValidateURLsPreserved(targetPath string, sourcePath string) ([]Violation, error) {
	var violations []Violation

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}
	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("read target: %w", err)
	}

	sourceURLs := urlPattern.FindAllString(string(sourceData), -1)
	targetURLs := urlPattern.FindAllString(string(targetData), -1)

	sourceSet := make(map[string]bool)
	for _, u := range sourceURLs {
		sourceSet[u] = true
	}
	for _, u := range targetURLs {
		delete(sourceSet, u)
	}
	for u := range sourceSet {
		violations = append(violations, Violation{Field: "urls", Message: fmt.Sprintf("URL missing: %s", u)})
	}

	return violations, nil
}
