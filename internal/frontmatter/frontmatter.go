package frontmatter

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Metadata struct {
	Title               string `yaml:"title"`
	TranslationKey      string `yaml:"translationKey"`
	Draft               bool   `yaml:"draft"`
	Reviewed            bool   `yaml:"reviewed"`
	SourceLang          string `yaml:"source_lang"`
	TargetLang          string `yaml:"target_lang"`
	TranslationProvider string `yaml:"translation_provider"`
	TranslationQuality  string `yaml:"translation_quality"`
}

type Document struct {
	Frontmatter string
	Body        string
	Metadata    Metadata
	RawMeta     map[string]any
}

func Split(markdown string) (Document, error) {
	if !strings.HasPrefix(markdown, "---\n") {
		return Document{Body: markdown}, nil
	}

	rest := strings.TrimPrefix(markdown, "---\n")
	parts := strings.SplitN(rest, "\n---\n", 2)
	if len(parts) != 2 {
		return Document{Body: markdown}, nil
	}

	fm := parts[0]
	var meta Metadata
	if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
		return Document{}, fmt.Errorf("parse frontmatter metadata: %w", err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(fm), &raw); err != nil {
		return Document{}, fmt.Errorf("parse frontmatter raw YAML: %w", err)
	}
	return Document{
		Frontmatter: fm,
		Body:        parts[1],
		Metadata:    meta,
		RawMeta:     raw,
	}, nil
}

type ProviderMeta struct {
	Provider string
	Quality  string
	Reviewed bool
	Draft    bool
}

func InjectProviderMeta(doc Document, pm ProviderMeta) (string, error) {
	if doc.RawMeta == nil {
		doc.RawMeta = make(map[string]any)
	}
	doc.RawMeta["translation_provider"] = pm.Provider
	doc.RawMeta["translation_quality"] = pm.Quality
	doc.RawMeta["reviewed"] = pm.Reviewed
	doc.RawMeta["draft"] = pm.Draft

	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(doc.RawMeta); err != nil {
		enc.Close()
		return "", fmt.Errorf("encode frontmatter: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("close YAML encoder: %w", err)
	}

	return "---\n" + strings.TrimSpace(b.String()) + "\n---\n" + doc.Body, nil
}
