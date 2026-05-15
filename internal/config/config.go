package config

type Config struct {
	Project     ProjectConfig     `yaml:"project" json:"project"`
	Paths       PathsConfig       `yaml:"paths" json:"paths"`
	Adapter     AdapterConfig     `yaml:"adapter" json:"adapter"`
	URLPolicy   URLPolicyConfig   `yaml:"url_policy" json:"url_policy"`
	Translation TranslationConfig `yaml:"translation" json:"translation"`
	Style       StyleConfig       `yaml:"style" json:"style"`
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
