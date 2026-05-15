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
	RawMeta     map[string]any
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
	var raw map[string]any
	_ = yaml.Unmarshal([]byte(fm), &raw)
	return Document{
		Frontmatter: fm,
		Body:        parts[1],
		Metadata:    meta,
		RawMeta:     raw,
	}
}

type ProviderMeta struct {
	Provider string
	Quality  string
	Reviewed bool
	Draft    bool
}

func InjectProviderMeta(doc Document, pm ProviderMeta) string {
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
	_ = enc.Encode(doc.RawMeta)
	enc.Close()

	return "---\n" + strings.TrimSpace(b.String()) + "\n---\n" + doc.Body
}
