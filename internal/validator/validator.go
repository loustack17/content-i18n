package validator

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/loustack17/content-i18n/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

type Violation struct {
	Field         string
	Section       string
	Message       string
	SuggestedFix  string
}

type ValidateOptions struct {
	GlossaryPath string
	BannedWords  []string
}

func Validate(targetPath string, sourcePath string, opts *ValidateOptions) ([]Violation, error) {
	var violations []Violation

	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("read target: %w", err)
	}

	targetContent := string(targetData)

	if !strings.HasPrefix(strings.TrimSpace(targetContent), "---") {
		violations = append(violations, Violation{Field: "frontmatter", Section: "header", Message: "target missing frontmatter", SuggestedFix: "add YAML frontmatter block"})
	}

	targetDoc := frontmatter.Split(targetContent)

	if targetDoc.Metadata.Title == "" {
		violations = append(violations, Violation{Field: "title", Section: "header", Message: "target title missing", SuggestedFix: "add title to frontmatter"})
	}

	if targetDoc.Metadata.Draft && !targetDoc.Metadata.Reviewed {
	} else if !targetDoc.Metadata.Draft && !targetDoc.Metadata.Reviewed {
		violations = append(violations, Violation{Field: "draft", Section: "header", Message: "draft should be true until reviewed", SuggestedFix: "set draft: true"})
	}

	cjkInTitle := countCJK(targetDoc.Metadata.Title)
	if cjkInTitle > 0 {
		violations = append(violations, Violation{Field: "title", Section: "header", Message: fmt.Sprintf("title contains %d CJK character(s)", cjkInTitle), SuggestedFix: "translate title to target language"})
	}

	cjkRatio := cjkRatioInBody(targetDoc.Body)
	if cjkRatio > 0.05 {
		violations = append(violations, Violation{Field: "language", Section: "body", Message: fmt.Sprintf("CJK ratio %.1f%% exceeds 5%% threshold", cjkRatio*100), SuggestedFix: "translate CJK content or wrap in code/quote"})
	}

	if targetDoc.Metadata.SourceLang == "" {
		violations = append(violations, Violation{Field: "language", Section: "header", Message: "source language metadata missing", SuggestedFix: "add source_lang to frontmatter"})
	}

	if targetDoc.Metadata.TargetLang == "" {
		violations = append(violations, Violation{Field: "language", Section: "header", Message: "target language metadata missing", SuggestedFix: "add target_lang to frontmatter"})
	}

	if !isTitleCapitalized(targetDoc.Metadata.Title) {
		violations = append(violations, Violation{Field: "title", Section: "header", Message: "title does not use natural English capitalization", SuggestedFix: "use title case for English titles"})
	}

	if sourcePath == "" {
		if opts != nil && len(opts.BannedWords) > 0 {
			for _, banned := range opts.BannedWords {
				if strings.Contains(strings.ToLower(targetContent), strings.ToLower(banned)) {
					violations = append(violations, Violation{Field: "style", Section: "body", Message: fmt.Sprintf("banned wording %q found in target", banned), SuggestedFix: "replace with approved terminology"})
				}
			}
		}

		return violations, nil
	}

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}

	sourceContent := string(sourceData)
	sourceDoc := frontmatter.Split(sourceContent)

	if sourceDoc.Metadata.TranslationKey != "" && targetDoc.Metadata.TranslationKey != sourceDoc.Metadata.TranslationKey {
		violations = append(violations, Violation{Field: "translationKey", Section: "header", Message: "translationKey mismatch", SuggestedFix: "set matching translationKey"})
	}

	if sourceDoc.Metadata.SourceLang != "" && targetDoc.Metadata.SourceLang != sourceDoc.Metadata.SourceLang {
		violations = append(violations, Violation{Field: "language", Section: "header", Message: fmt.Sprintf("source_lang mismatch: source has %q, target has %q", sourceDoc.Metadata.SourceLang, targetDoc.Metadata.SourceLang), SuggestedFix: "set matching source_lang"})
	}

	sourceCodeBlocks := frontmatter.ExtractCodeBlocks(sourceContent)
	targetCodeBlocks := frontmatter.ExtractCodeBlocks(targetContent)
	if len(targetCodeBlocks) != len(sourceCodeBlocks) {
		violations = append(violations, Violation{Field: "codeBlocks", Section: "body", Message: fmt.Sprintf("target has %d code blocks, source has %d", len(targetCodeBlocks), len(sourceCodeBlocks)), SuggestedFix: "preserve all code blocks from source"})
	} else {
		for i := range sourceCodeBlocks {
			if strings.TrimSpace(sourceCodeBlocks[i]) != strings.TrimSpace(targetCodeBlocks[i]) {
				violations = append(violations, Violation{Field: "codeBlocks", Section: "body", Message: fmt.Sprintf("code block %d content differs from source", i+1), SuggestedFix: "restore original code block"})
				break
			}
		}
	}

	sourceInline := frontmatter.ExtractInlineCode(sourceContent)
	targetInline := frontmatter.ExtractInlineCode(targetContent)
	if len(targetInline) < len(sourceInline) {
		violations = append(violations, Violation{Field: "inlineCode", Section: "body", Message: fmt.Sprintf("target has %d inline code, source has %d", len(targetInline), len(sourceInline)), SuggestedFix: "preserve all inline code from source"})
	}

	if len(sourceInline) == len(targetInline) {
		for i := range sourceInline {
			if strings.TrimSpace(sourceInline[i]) != strings.TrimSpace(targetInline[i]) {
				violations = append(violations, Violation{Field: "inlineCode", Section: "body", Message: fmt.Sprintf("inline code %q changed to %q", sourceInline[i], targetInline[i]), SuggestedFix: "restore original inline code"})
				break
			}
		}
	}

	sourceURLs := urlPattern.FindAllString(sourceContent, -1)
	targetURLs := urlPattern.FindAllString(targetContent, -1)
	sourceURLSet := make(map[string]bool)
	for _, u := range sourceURLs {
		sourceURLSet[u] = true
	}
	for _, u := range targetURLs {
		delete(sourceURLSet, u)
	}
	for u := range sourceURLSet {
		violations = append(violations, Violation{Field: "urls", Section: "body", Message: fmt.Sprintf("URL missing: %s", u), SuggestedFix: "restore URL from source"})
	}

	if opts != nil && opts.GlossaryPath != "" {
		glossaryTerms, err := loadGlossaryTerms(opts.GlossaryPath)
		if err == nil {
			for _, term := range glossaryTerms {
				if strings.Contains(sourceContent, term.Source) && !strings.Contains(targetContent, term.Target) {
					violations = append(violations, Violation{Field: "glossary", Section: "body", Message: fmt.Sprintf("glossary term %q -> %q not found in target", term.Source, term.Target), SuggestedFix: "add glossary term to translation"})
				}
			}
		}
	}

	if opts != nil && len(opts.BannedWords) > 0 {
		for _, banned := range opts.BannedWords {
			if strings.Contains(strings.ToLower(targetContent), strings.ToLower(banned)) {
				violations = append(violations, Violation{Field: "style", Section: "body", Message: fmt.Sprintf("banned wording %q found in target", banned), SuggestedFix: "replace with approved terminology"})
			}
		}
	}

	return violations, nil
}

type GlossaryTerm struct {
	Source string
	Target string
}

type GlossaryFile struct {
	Terms []struct {
		Source string `yaml:"source"`
		Target string `yaml:"target"`
	} `yaml:"terms"`
}

func loadGlossaryTerms(path string) ([]GlossaryTerm, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		var gf GlossaryFile
		if err := yaml.Unmarshal(data, &gf); err != nil {
			return nil, fmt.Errorf("parse glossary YAML: %w", err)
		}
		terms := make([]GlossaryTerm, 0, len(gf.Terms))
		for _, t := range gf.Terms {
			terms = append(terms, GlossaryTerm{Source: t.Source, Target: t.Target})
		}
		return terms, nil
	}

	var terms []GlossaryTerm
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 {
			terms = append(terms, GlossaryTerm{
				Source: strings.TrimSpace(parts[0]),
				Target: strings.TrimSpace(parts[1]),
			})
		}
	}
	return terms, nil
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

func isTitleCapitalized(title string) bool {
	if title == "" {
		return true
	}
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) == 0 {
			continue
		}
		runes := []rune(word)
		if i == 0 {
			if !unicode.IsUpper(runes[0]) {
				return false
			}
		} else {
			if len(word) <= 3 {
				continue
			}
			if !unicode.IsUpper(runes[0]) {
				return false
			}
		}
	}
	return true
}

var urlPattern = regexp.MustCompile(`https?://[^\s\)\]"<>]+`)
