package frontmatter

import (
	"bytes"
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
}

func Split(markdown string) Document {
	if !strings.HasPrefix(markdown, "---\n") {
		return Document{Body: markdown}
	}

	rest := strings.TrimPrefix(markdown, "---\n")
	parts := strings.SplitN(rest, "\n---\n", 2)
	if len(parts) != 2 {
		return Document{Body: markdown}
	}

	fm := parts[0]
	var meta Metadata
	_ = yaml.Unmarshal([]byte(fm), &meta)
	return Document{
		Frontmatter: fm,
		Body:        parts[1],
		Metadata:    meta,
	}
}

type ProviderMeta struct {
	Provider string
	Quality  string
	Reviewed bool
	Draft    bool
}

func InjectProviderMeta(doc Document, pm ProviderMeta) string {
	meta := doc.Metadata
	meta.TranslationProvider = pm.Provider
	meta.TranslationQuality = pm.Quality
	meta.Reviewed = pm.Reviewed
	meta.Draft = pm.Draft

	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	_ = enc.Encode(meta)
	enc.Close()

	return "---\n" + strings.TrimSpace(b.String()) + "\n---\n" + doc.Body
}
