package core

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
)

type WorkPacket struct {
	Dir          string
	SourcePath   string
	TargetPath   string
	PromptPath   string
	GlossaryPath string
	StylePath    string
	MetaPath     string
}

type WorkMeta struct {
	SourcePath     string               `json:"source_path"`
	TargetLanguage string               `json:"target_language"`
	Provider       string               `json:"provider,omitempty"`
	StructureHash  string               `json:"structure_hash"`
	Fingerprint    StructureFingerprint `json:"fingerprint"`
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

	fp := computeFingerprint(string(sourceData))
	meta := WorkMeta{
		SourcePath:     sourceFile,
		TargetLanguage: targetLang,
		Provider:       "manual",
		StructureHash:  fp.Hash,
		Fingerprint:    fp.Fingerprint,
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal meta: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "meta.json"), metaData, 0644); err != nil {
		return nil, fmt.Errorf("write meta: %w", err)
	}

	return &WorkPacket{
		Dir:          workDir,
		SourcePath:   filepath.Join(workDir, "source.md"),
		TargetPath:   targetPath,
		PromptPath:   filepath.Join(workDir, "prompt.md"),
		GlossaryPath: filepath.Join(workDir, "glossary.md"),
		StylePath:    filepath.Join(workDir, "style.md"),
		MetaPath:     filepath.Join(workDir, "meta.json"),
	}, nil
}

type StructureFingerprint struct {
	HeadingCount       int `json:"heading_count"`
	H2Count            int `json:"h2_count"`
	H3Count            int `json:"h3_count"`
	H4Count            int `json:"h4_count"`
	OrderedListCount   int `json:"ordered_list_count"`
	UnorderedListCount int `json:"unordered_list_count"`
	TableCount         int `json:"table_count"`
	ParagraphCount     int `json:"paragraph_count"`
	BlockquoteCount    int `json:"blockquote_count"`
	CodeBlockCount     int `json:"code_block_count"`
}

type FingerprintResult struct {
	Fingerprint StructureFingerprint
	Hash        string
}

var (
	h2Re    = regexp.MustCompile(`(?m)^## `)
	h3Re    = regexp.MustCompile(`(?m)^### `)
	h4Re    = regexp.MustCompile(`(?m)^#### `)
	olRe    = regexp.MustCompile(`(?m)^\d+\.\s`)
	ulRe    = regexp.MustCompile(`(?m)^[-*+]\s`)
	tableRe = regexp.MustCompile(`(?m)^\|`)
	bqRe    = regexp.MustCompile(`(?m)^> `)
	fenceRe = regexp.MustCompile("(?m)^```")
)

func countParagraphs(body string) int {
	lines := strings.Split(body, "\n")
	count := 0
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inBlock {
				count++
				inBlock = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, ">") {
			if inBlock {
				count++
				inBlock = false
			}
			continue
		}
		inBlock = true
	}
	if inBlock {
		count++
	}
	return count
}

func computeFingerprint(markdown string) FingerprintResult {
	body := markdown
	if idx := strings.Index(markdown, "---\n"); idx >= 0 {
		rest := markdown[idx+4:]
		if endIdx := strings.Index(rest, "\n---\n"); endIdx >= 0 {
			body = rest[endIdx+5:]
		}
	}

	fp := StructureFingerprint{
		HeadingCount:       len(h2Re.FindAllString(markdown, -1)) + len(h3Re.FindAllString(markdown, -1)) + len(h4Re.FindAllString(markdown, -1)),
		H2Count:            len(h2Re.FindAllString(markdown, -1)),
		H3Count:            len(h3Re.FindAllString(markdown, -1)),
		H4Count:            len(h4Re.FindAllString(markdown, -1)),
		OrderedListCount:   len(olRe.FindAllString(body, -1)),
		UnorderedListCount: len(ulRe.FindAllString(body, -1)),
		TableCount:         len(tableRe.FindAllString(body, -1)),
		ParagraphCount:     countParagraphs(body),
		BlockquoteCount:    len(bqRe.FindAllString(body, -1)),
		CodeBlockCount:     len(fenceRe.FindAllString(markdown, -1)) / 2,
	}

	data, _ := json.Marshal(fp)
	h := sha256.Sum256(data)
	return FingerprintResult{Fingerprint: fp, Hash: fmt.Sprintf("%x", h[:8])}
}
