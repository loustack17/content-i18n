package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
	"github.com/loustack17/content-i18n/internal/structure"
)

type WorkPacket struct {
	Dir          string
	SourcePath   string
	TargetPath   string
	PromptPath   string
	GlossaryPath string
	StylePath    string
	ContextPath  string
	MetaPath     string
}

type WorkMeta struct {
	SourcePath     string                         `json:"source_path"`
	TargetLanguage string                         `json:"target_language"`
	Provider       string                         `json:"provider,omitempty"`
	StructureHash  string                         `json:"structure_hash"`
	Fingerprint    structure.StructureFingerprint `json:"fingerprint"`
	Headings       []string                       `json:"headings"`
	URLs           []string                       `json:"urls"`
}

func SlugFromPath(sourcePath string, sourceRoot string) string {
	rel, err := filepath.Rel(sourceRoot, sourcePath)
	if err != nil {
		rel = sourcePath
	}
	rel = strings.TrimSuffix(rel, ".md")
	return strings.ReplaceAll(rel, string(filepath.Separator), "-")
}

func GenerateWorkPacket(cfg *config.Config, sourceFile string, targetLang string) (*WorkPacket, error) {
	slug := SlugFromPath(sourceFile, cfg.Paths.Source)
	workDir := filepath.Join("work", slug)

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("create work dir: %w", err)
	}

	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "source.md"), sourceData, 0644); err != nil {
		return nil, err
	}

	promptPath := filepath.Join(cfg.ConfigDir, "prompts", "translate-section.md")
	if _, err := os.Stat(promptPath); err == nil {
		promptData, err := os.ReadFile(promptPath)
		if err != nil {
			return nil, fmt.Errorf("read prompt: %w", err)
		}
		if err := os.WriteFile(filepath.Join(workDir, "prompt.md"), promptData, 0644); err != nil {
			return nil, fmt.Errorf("write prompt: %w", err)
		}
	}

	if cfg.Style.Glossary != "" {
		glossaryPath := cfg.Style.Glossary
		if !filepath.IsAbs(glossaryPath) {
			glossaryPath = filepath.Join(cfg.ConfigDir, glossaryPath)
		}
		if _, err := os.Stat(glossaryPath); err == nil {
			glossaryData, err := os.ReadFile(glossaryPath)
			if err != nil {
				return nil, fmt.Errorf("read glossary: %w", err)
			}
			if err := os.WriteFile(filepath.Join(workDir, "glossary.md"), glossaryData, 0644); err != nil {
				return nil, fmt.Errorf("write glossary: %w", err)
			}
		}
	}

	if cfg.Style.Pack != "" {
		stylePath := cfg.Style.Pack
		if !filepath.IsAbs(stylePath) {
			stylePath = filepath.Join(cfg.ConfigDir, stylePath)
		}
		if _, err := os.Stat(stylePath); err == nil {
			styleData, err := os.ReadFile(stylePath)
			if err != nil {
				return nil, fmt.Errorf("read style: %w", err)
			}
			if err := os.WriteFile(filepath.Join(workDir, "style.md"), styleData, 0644); err != nil {
				return nil, fmt.Errorf("write style: %w", err)
			}
		}
	}

	targetPath := filepath.Join(workDir, "target.md")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		if err := os.WriteFile(targetPath, []byte{}, 0644); err != nil {
			return nil, fmt.Errorf("create target: %w", err)
		}
	}

	sourceText := string(sourceData)
	fp := structure.ComputeFingerprint(sourceText)
	headings := structure.ExtractHeadings(sourceText)
	urls := structure.UniqueStrings(content.URLPattern.FindAllString(sourceText, -1))
	meta := WorkMeta{
		SourcePath:     sourceFile,
		TargetLanguage: targetLang,
		Provider:       "manual",
		StructureHash:  fp.Hash,
		Fingerprint:    fp.Fingerprint,
		Headings:       headings,
		URLs:           urls,
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal meta: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "meta.json"), metaData, 0644); err != nil {
		return nil, fmt.Errorf("write meta: %w", err)
	}

	contextData := buildHarnessContext(meta)
	if err := os.WriteFile(filepath.Join(workDir, "context.md"), []byte(contextData), 0644); err != nil {
		return nil, fmt.Errorf("write context: %w", err)
	}

	return &WorkPacket{
		Dir:          workDir,
		SourcePath:   filepath.Join(workDir, "source.md"),
		TargetPath:   targetPath,
		PromptPath:   filepath.Join(workDir, "prompt.md"),
		GlossaryPath: filepath.Join(workDir, "glossary.md"),
		StylePath:    filepath.Join(workDir, "style.md"),
		ContextPath:  filepath.Join(workDir, "context.md"),
		MetaPath:     filepath.Join(workDir, "meta.json"),
	}, nil
}

func buildHarnessContext(meta WorkMeta) string {
	var b strings.Builder
	b.WriteString("# Translation Harness Context\n\n")
	b.WriteString("## Hard Contract\n\n")
	b.WriteString("- Translate language only.\n")
	b.WriteString("- Preserve structure, content coverage, argument flow, and style class.\n")
	b.WriteString("- Do not summarize, compress, merge, split, or editorialize.\n")
	b.WriteString("- The English target must be the same article in another language.\n\n")
	b.WriteString("## Source Structure Fingerprint\n\n")
	fmt.Fprintf(&b, "- structure_hash: `%s`\n", meta.StructureHash)
	fmt.Fprintf(&b, "- H2 headings: %d\n", meta.Fingerprint.H2Count)
	fmt.Fprintf(&b, "- H3 headings: %d\n", meta.Fingerprint.H3Count)
	fmt.Fprintf(&b, "- H4 headings: %d\n", meta.Fingerprint.H4Count)
	fmt.Fprintf(&b, "- ordered list items: %d\n", meta.Fingerprint.OrderedListCount)
	fmt.Fprintf(&b, "- unordered list items: %d\n", meta.Fingerprint.UnorderedListCount)
	fmt.Fprintf(&b, "- table rows: %d\n", meta.Fingerprint.TableCount)
	fmt.Fprintf(&b, "- paragraphs: %d\n", meta.Fingerprint.ParagraphCount)
	fmt.Fprintf(&b, "- blockquotes: %d\n", meta.Fingerprint.BlockquoteCount)
	fmt.Fprintf(&b, "- fenced code blocks: %d\n\n", meta.Fingerprint.CodeBlockCount)
	b.WriteString("## Heading Order\n\n")
	for i, h := range meta.Headings {
		fmt.Fprintf(&b, "%d. %s\n", i+1, h)
	}
	b.WriteString("\n## Preserved URLs\n\n")
	if len(meta.URLs) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, u := range meta.URLs {
			fmt.Fprintf(&b, "- %s\n", u)
		}
	}
	b.WriteString("\n## Translation Self-Check\n\n")
	b.WriteString("1. Heading hierarchy and order match source.\n")
	b.WriteString("2. Paragraph count per section matches source.\n")
	b.WriteString("3. Lists, tables, examples, references, blockquotes, and URLs remain present.\n")
	b.WriteString("4. Code blocks, inline code, commands, and identifiers remain exact.\n")
	b.WriteString("5. No Mandarin prose remains in the English target.\n")
	b.WriteString("6. No section is shortened into a summary.\n")
	b.WriteString("7. Style class matches source.\n")
	return b.String()
}
