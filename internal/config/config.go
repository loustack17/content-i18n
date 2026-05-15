package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project     ProjectConfig     `yaml:"project" json:"project"`
	Paths       PathsConfig       `yaml:"paths" json:"paths"`
	Adapter     AdapterConfig     `yaml:"adapter" json:"adapter"`
	URLPolicy   URLPolicyConfig   `yaml:"url_policy" json:"url_policy"`
	Translation TranslationConfig `yaml:"translation" json:"translation"`
	Style       StyleConfig       `yaml:"style" json:"style"`
	ConfigDir   string
}

type ProjectConfig struct {
	Type            string   `yaml:"type" json:"type"`
	SourceLanguage  string   `yaml:"source_language" json:"source_language"`
	TargetLanguages []string `yaml:"target_languages" json:"target_languages"`
}

type PathsConfig struct {
	Source  string            `yaml:"source" json:"source"`
	Targets map[string]string `yaml:"targets" json:"targets"`
}

type AdapterConfig struct {
	Name                  string `yaml:"name" json:"name"`
	Mode                  string `yaml:"mode" json:"mode"`
	PreserveRelativePaths bool   `yaml:"preserve_relative_paths" json:"preserve_relative_paths"`
	TranslationKey        string `yaml:"translation_key" json:"translation_key"`
}

type URLPolicyConfig struct {
	Canonical          map[string]string        `yaml:"canonical" json:"canonical"`
	RuntimeTranslation RuntimeTranslationConfig `yaml:"runtime_translation" json:"runtime_translation"`
}

type RuntimeTranslationConfig struct {
	Enabled            bool     `yaml:"enabled" json:"enabled"`
	Route              string   `yaml:"route" json:"route"`
	QueryParam         string   `yaml:"query_param" json:"query_param"`
	CanonicalLanguages []string `yaml:"canonical_languages" json:"canonical_languages"`
}

type TranslationConfig struct {
	DefaultProvider   string       `yaml:"default_provider" json:"default_provider"`
	FallbackProviders []string     `yaml:"fallback_providers" json:"fallback_providers"`
	Output            OutputConfig `yaml:"output" json:"output"`
}

type OutputConfig struct {
	Draft                   bool `yaml:"draft" json:"draft"`
	Reviewed                bool `yaml:"reviewed" json:"reviewed"`
	PreserveCodeBlocks      bool `yaml:"preserve_code_blocks" json:"preserve_code_blocks"`
	PreserveInlineCode      bool `yaml:"preserve_inline_code" json:"preserve_inline_code"`
	PreserveFrontmatterKeys bool `yaml:"preserve_frontmatter_keys" json:"preserve_frontmatter_keys"`
	PreserveLinks           bool `yaml:"preserve_links" json:"preserve_links"`
}

type StyleConfig struct {
	Pack     string `yaml:"pack" json:"pack"`
	Glossary string `yaml:"glossary" json:"glossary"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfgDir := filepath.Dir(path)
	if err := resolvePaths(&cfg, cfgDir); err != nil {
		return nil, err
	}
	cfg.ConfigDir = cfgDir

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) StatusFilePath() string {
	return filepath.Join(c.ConfigDir, ".content-i18n", "status.json")
}

func resolvePaths(cfg *Config, cfgDir string) error {
	if !filepath.IsAbs(cfg.Paths.Source) {
		cfg.Paths.Source = filepath.Join(cfgDir, cfg.Paths.Source)
	}
	for lang, target := range cfg.Paths.Targets {
		if !filepath.IsAbs(target) {
			cfg.Paths.Targets[lang] = filepath.Join(cfgDir, target)
		}
	}
	return nil
}

func validate(cfg *Config) error {
	if cfg.Project.Type == "" {
		return fmt.Errorf("project.type required")
	}
	if cfg.Project.SourceLanguage == "" {
		return fmt.Errorf("project.source_language required")
	}
	if len(cfg.Project.TargetLanguages) == 0 {
		return fmt.Errorf("project.target_languages required")
	}
	if cfg.Paths.Source == "" {
		return fmt.Errorf("paths.source required")
	}
	if len(cfg.Paths.Targets) == 0 {
		return fmt.Errorf("paths.targets required")
	}
	if cfg.Translation.DefaultProvider == "" {
		return fmt.Errorf("translation.default_provider required")
	}

	if _, err := os.Stat(cfg.Paths.Source); err != nil {
		return fmt.Errorf("paths.source not found: %w", err)
	}

	for lang, target := range cfg.Paths.Targets {
		dir := filepath.Dir(target)
		if _, err := os.Stat(dir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("paths.targets[%s] error: %w", lang, err)
		}
	}

	return nil
}
