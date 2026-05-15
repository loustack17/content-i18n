package frontmatter

import (
	"strings"
)

type Section struct {
	Level   int
	Heading string
	Body    string
}

func SplitSections(body string) []Section {
	lines := strings.Split(body, "\n")
	var sections []Section
	var current Section
	var buffer []string

	flush := func() {
		if len(buffer) > 0 {
			current.Body = strings.Join(buffer, "\n")
			sections = append(sections, current)
			buffer = nil
		}
	}

	for _, line := range lines {
		level := headingLevel(line)
		if level > 0 {
			flush()
			current = Section{Level: level, Heading: strings.TrimSpace(strings.TrimLeft(line, "# "))}
		} else {
			buffer = append(buffer, line)
		}
	}
	flush()

	return sections
}

func headingLevel(line string) int {
	for i, c := range line {
		if c != '#' {
			if i > 0 && i <= 6 && (i == len(line) || line[i] == ' ') {
				return i
			}
			return 0
		}
	}
	return 0
}

func ExtractCodeBlocks(body string) []string {
	var blocks []string
	lines := strings.Split(body, "\n")
	var inBlock bool
	var block []string
	var fence string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inBlock && strings.HasPrefix(trimmed, "```") {
			inBlock = true
			fence = trimmed
			block = []string{line}
			continue
		}
		if inBlock {
			block = append(block, line)
			trimmed := strings.TrimSpace(line)
			if trimmed == fence || (strings.HasPrefix(trimmed, "```") && len(trimmed) == 3) {
				blocks = append(blocks, strings.Join(block, "\n"))
				inBlock = false
				block = nil
			}
		}
	}

	return blocks
}

func ExtractInlineCode(body string) []string {
	var spans []string
	var inCode bool
	var span strings.Builder

	for i := 0; i < len(body); i++ {
		if body[i] == '`' && (i == 0 || body[i-1] != '\\') {
			if inCode {
				spans = append(spans, span.String())
				span.Reset()
			}
			inCode = !inCode
			continue
		}
		if inCode {
			span.WriteByte(body[i])
		}
	}

	return spans
}
