package frontmatter

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type Metadata struct {
	Title          string `yaml:"title"`
	TranslationKey string `yaml:"translationKey"`
	Draft          bool   `yaml:"draft"`
	Reviewed       bool   `yaml:"reviewed"`
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
