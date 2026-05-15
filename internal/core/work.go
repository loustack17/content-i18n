package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	SourcePath     string `json:"source_path"`
	TargetLanguage string `json:"target_language"`
	Provider       string `json:"provider,omitempty"`
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

	meta := WorkMeta{
		SourcePath:     sourceFile,
		TargetLanguage: targetLang,
		Provider:       "manual",
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
