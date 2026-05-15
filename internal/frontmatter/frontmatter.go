package frontmatter

import "strings"

type Document struct {
	Frontmatter string
	Body        string
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

	return Document{Frontmatter: parts[0], Body: parts[1]}
}
