package validator

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/loustack17/content-i18n/internal/content"
	"github.com/loustack17/content-i18n/internal/frontmatter"
	"github.com/loustack17/content-i18n/internal/structure"
	"gopkg.in/yaml.v3"
)

type ValidateOptions struct {
	GlossaryPath string
	BannedWords  []string
	ToneChecks   ToneCheckOptions
}

type ToneCheckOptions struct {
	AbstractOpenerThreshold int
	AbstractTerms           []string
	HeadingDocLikePrefixes  []string
}

type Violation = structure.Violation

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

	targetDoc, err := frontmatter.Split(targetContent)
	if err != nil {
		violations = append(violations, Violation{Field: "frontmatter", Section: "header", Message: fmt.Sprintf("invalid frontmatter: %v", err), SuggestedFix: "fix YAML frontmatter syntax"})
		return violations, nil
	}

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

		if opts != nil {
			violations = append(violations, checkTone(targetDoc.Body, opts.ToneChecks, opts.GlossaryPath)...)
		}

		return violations, nil
	}

	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}

	sourceContent := string(sourceData)
	sourceDoc, err := frontmatter.Split(sourceContent)
	if err != nil {
		return nil, fmt.Errorf("parse source frontmatter: %w", err)
	}

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

	sourceURLs := content.URLPattern.FindAllString(sourceContent, -1)
	targetURLs := content.URLPattern.FindAllString(targetContent, -1)
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

	violations = append(violations, checkStructure(sourceContent, targetContent)...)

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

	if opts != nil {
		violations = append(violations, checkTone(targetDoc.Body, opts.ToneChecks, opts.GlossaryPath)...)
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

func checkTone(body string, opts ToneCheckOptions, glossaryPath string) []Violation {
	var violations []Violation

	if opts.AbstractOpenerThreshold > 0 {
		violations = append(violations, checkAbstractOpeners(body, opts.AbstractOpenerThreshold)...)
	}

	if len(opts.AbstractTerms) > 0 {
		violations = append(violations, checkAbstractTermOveruse(body, opts.AbstractTerms, glossaryPath)...)
	}

	if len(opts.HeadingDocLikePrefixes) > 0 {
		violations = append(violations, checkHeadingPhrasing(body, opts.HeadingDocLikePrefixes)...)
	}

	return violations
}

var abstractOpenerRe = regexp.MustCompile(`(?m)^(?:The|This|These|Those|A|An)\s+\w+\s+(?:is|are|was|were|has|have)\b`)

func checkAbstractOpeners(body string, threshold int) []Violation {
	matches := abstractOpenerRe.FindAllString(body, -1)
	if len(matches) > threshold {
		return []Violation{{
			Field:        "tone",
			Section:      "body",
			Message:      fmt.Sprintf("%d repeated abstract openers (e.g. 'The ... is ...') exceed threshold %d", len(matches), threshold),
			SuggestedFix: "vary sentence openers; use active voice or concrete subjects",
		}}
	}
	return nil
}

func checkAbstractTermOveruse(body string, abstractTerms []string, glossaryPath string) []Violation {
	glossaryTargets := make(map[string]bool)
	if glossaryPath != "" {
		terms, err := loadGlossaryTerms(glossaryPath)
		if err == nil {
			for _, t := range terms {
				glossaryTargets[strings.ToLower(t.Target)] = true
			}
		}
	}

	bodyLower := strings.ToLower(body)
	for _, term := range abstractTerms {
		termLower := strings.ToLower(term)
		count := strings.Count(bodyLower, termLower)
		if count > 2 && !glossaryTargets[termLower] {
			return []Violation{{
				Field:        "tone",
				Section:      "body",
				Message:      fmt.Sprintf("abstract term %q appears %d times; prefer concrete glossary terms", term, count),
				SuggestedFix: "replace with concrete terminology or approved glossary equivalent",
			}}
		}
	}
	return nil
}

var headingRe = regexp.MustCompile(`(?m)^#{1,6}\s+(.*)`)

func checkHeadingPhrasing(body string, docLikePrefixes []string) []Violation {
	headings := headingRe.FindAllStringSubmatch(body, -1)
	var violations []Violation
	for _, h := range headings {
		if len(h) < 2 {
			continue
		}
		headingText := strings.TrimSpace(h[1])
		for _, prefix := range docLikePrefixes {
			if strings.HasPrefix(strings.ToLower(headingText), prefix) {
				violations = append(violations, Violation{
					Field:        "tone",
					Section:      "heading",
					Message:      fmt.Sprintf("heading %q uses doc-like prefix %q", headingText, prefix),
					SuggestedFix: "use blog-style action/outcome headings instead of documentation labels",
				})
			}
		}
	}
	return violations
}

func checkStructure(source, target string) []Violation {
	var violations []Violation

	srcBody := structure.ExtractBody(source)
	tgtBody := structure.ExtractBody(target)

	srcH2 := len(structure.H2Re.FindAllString(source, -1))
	tgtH2 := len(structure.H2Re.FindAllString(target, -1))
	if srcH2 != tgtH2 {
		violations = append(violations, Violation{Field: "structure", Section: "headings", Message: fmt.Sprintf("target has %d H2 headings, source has %d", tgtH2, srcH2), SuggestedFix: "preserve heading hierarchy from source"})
	}

	srcH3 := len(structure.H3Re.FindAllString(source, -1))
	tgtH3 := len(structure.H3Re.FindAllString(target, -1))
	if srcH3 != tgtH3 {
		violations = append(violations, Violation{Field: "structure", Section: "headings", Message: fmt.Sprintf("target has %d H3 headings, source has %d", tgtH3, srcH3), SuggestedFix: "preserve heading hierarchy from source"})
	}

	srcH4 := len(structure.H4Re.FindAllString(source, -1))
	tgtH4 := len(structure.H4Re.FindAllString(target, -1))
	if srcH4 != tgtH4 {
		violations = append(violations, Violation{Field: "structure", Section: "headings", Message: fmt.Sprintf("target has %d H4 headings, source has %d", tgtH4, srcH4), SuggestedFix: "preserve heading hierarchy from source"})
	}

	srcHeadings := structure.ExtractHeadings(source)
	tgtHeadings := structure.ExtractHeadings(target)
	if len(srcHeadings) == len(tgtHeadings) {
		for i := range srcHeadings {
			srcNorm := structure.NormalizeHeadingText(srcHeadings[i])
			tgtNorm := structure.NormalizeHeadingText(tgtHeadings[i])
			if srcNorm == tgtNorm {
				continue
			}
			srcWords := meaningfulHeadingWords(srcNorm)
			tgtWords := meaningfulHeadingWords(tgtNorm)
			common := 0
			for _, sw := range srcWords {
				for _, tw := range tgtWords {
					if strings.EqualFold(sw, tw) {
						common++
						break
					}
				}
			}
			if common == 0 {
				continue
			}
			violations = append(violations, Violation{Field: "structure", Section: "headings", Message: fmt.Sprintf("heading order mismatch at position %d: source has %q, target has %q", i+1, srcHeadings[i], tgtHeadings[i]), SuggestedFix: "preserve section order from source"})
			break
		}
	}

	srcOL := len(structure.OLRe.FindAllString(srcBody, -1))
	tgtOL := len(structure.OLRe.FindAllString(tgtBody, -1))
	if srcOL != tgtOL {
		violations = append(violations, Violation{Field: "structure", Section: "lists", Message: fmt.Sprintf("target has %d ordered list items, source has %d", tgtOL, srcOL), SuggestedFix: "preserve list structure from source"})
	}

	srcUL := len(structure.ULRe.FindAllString(srcBody, -1))
	tgtUL := len(structure.ULRe.FindAllString(tgtBody, -1))
	if srcUL != tgtUL {
		violations = append(violations, Violation{Field: "structure", Section: "lists", Message: fmt.Sprintf("target has %d unordered list items, source has %d", tgtUL, srcUL), SuggestedFix: "preserve list structure from source"})
	}

	srcTables := len(structure.TableRe.FindAllString(srcBody, -1))
	tgtTables := len(structure.TableRe.FindAllString(tgtBody, -1))
	if srcTables != tgtTables {
		violations = append(violations, Violation{Field: "structure", Section: "tables", Message: fmt.Sprintf("target has %d table rows, source has %d", tgtTables, srcTables), SuggestedFix: "preserve table structure from source"})
	}

	srcTableCols := structure.CountTableColumns(srcBody)
	tgtTableCols := structure.CountTableColumns(tgtBody)
	if len(srcTableCols) == len(tgtTableCols) {
		for i := range srcTableCols {
			if srcTableCols[i] != tgtTableCols[i] {
				violations = append(violations, Violation{Field: "structure", Section: "tables", Message: fmt.Sprintf("table %d has %d columns in target, %d in source", i+1, tgtTableCols[i], srcTableCols[i]), SuggestedFix: "preserve table column count from source"})
				break
			}
		}
	}

	srcBQ := len(structure.BQRe.FindAllString(srcBody, -1))
	tgtBQ := len(structure.BQRe.FindAllString(tgtBody, -1))
	if srcBQ != tgtBQ {
		violations = append(violations, Violation{Field: "structure", Section: "blockquotes", Message: fmt.Sprintf("target has %d blockquotes, source has %d", tgtBQ, srcBQ), SuggestedFix: "preserve blockquote structure from source"})
	}

	srcFences := len(structure.FenceRe.FindAllString(source, -1)) / 2
	tgtFences := len(structure.FenceRe.FindAllString(target, -1)) / 2
	if srcFences != tgtFences {
		violations = append(violations, Violation{Field: "structure", Section: "code", Message: fmt.Sprintf("target has %d fenced code blocks, source has %d", tgtFences, srcFences), SuggestedFix: "preserve all code blocks from source"})
	}

	srcParas := structure.CountParagraphs(srcBody)
	tgtParas := structure.CountParagraphs(tgtBody)
	if srcParas != tgtParas {
		violations = append(violations, Violation{Field: "structure", Section: "paragraphs", Message: fmt.Sprintf("target has %d paragraphs, source has %d", tgtParas, srcParas), SuggestedFix: "preserve paragraph count from source; do not merge or split paragraphs"})
	}

	violations = append(violations, structure.CheckOmission(srcBody, tgtBody)...)

	return violations
}

func meaningfulHeadingWords(heading string) []string {
	stripped := stripInlineCode(heading)
	tokens := strings.Fields(stripped)
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if isPunctuationOnly(t) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func stripInlineCode(heading string) string {
	result := heading
	for {
		start := strings.Index(result, "`")
		if start < 0 {
			break
		}
		end := strings.Index(result[start+1:], "`")
		if end < 0 {
			break
		}
		end = start + 1 + end
		result = result[:start] + " " + result[end+1:]
	}
	return result
}

func isPunctuationOnly(token string) bool {
	for _, r := range token {
		if !unicode.IsPunct(r) && !unicode.IsSymbol(r) {
			return false
		}
	}
	return true
}
