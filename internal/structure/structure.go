package structure

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

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
	H2Re      = regexp.MustCompile(`(?m)^## `)
	H3Re      = regexp.MustCompile(`(?m)^### `)
	H4Re      = regexp.MustCompile(`(?m)^#### `)
	OLRe      = regexp.MustCompile(`(?m)^\d+\.\s`)
	ULRe      = regexp.MustCompile(`(?m)^[-*+]\s`)
	TableRe   = regexp.MustCompile(`(?m)^\|`)
	BQRe      = regexp.MustCompile(`(?m)^> `)
	FenceRe   = regexp.MustCompile("(?m)^```")
	HeadingRe = regexp.MustCompile(`(?m)^(#{1,6})\s+(.*)`)
)

func ExtractBody(markdown string) string {
	if !strings.HasPrefix(markdown, "---\n") {
		return markdown
	}
	rest := markdown[4:]
	if endIdx := strings.Index(rest, "\n---\n"); endIdx >= 0 {
		return rest[endIdx+5:]
	}
	return markdown
}

func CountParagraphs(body string) int {
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

func ComputeFingerprint(markdown string) FingerprintResult {
	body := ExtractBody(markdown)

	fp := StructureFingerprint{
		HeadingCount:       len(H2Re.FindAllString(markdown, -1)) + len(H3Re.FindAllString(markdown, -1)) + len(H4Re.FindAllString(markdown, -1)),
		H2Count:            len(H2Re.FindAllString(markdown, -1)),
		H3Count:            len(H3Re.FindAllString(markdown, -1)),
		H4Count:            len(H4Re.FindAllString(markdown, -1)),
		OrderedListCount:   len(OLRe.FindAllString(body, -1)),
		UnorderedListCount: len(ULRe.FindAllString(body, -1)),
		TableCount:         len(TableRe.FindAllString(body, -1)),
		ParagraphCount:     CountParagraphs(body),
		BlockquoteCount:    len(BQRe.FindAllString(body, -1)),
		CodeBlockCount:     len(FenceRe.FindAllString(markdown, -1)) / 2,
	}

	data, err := json.Marshal(fp)
	if err != nil {
		panic(fmt.Sprintf("marshal fingerprint: %v", err))
	}
	h := sha256.Sum256(data)
	return FingerprintResult{Fingerprint: fp, Hash: fmt.Sprintf("%x", h[:8])}
}

func ExtractHeadings(markdown string) []string {
	matches := HeadingRe.FindAllStringSubmatch(markdown, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) >= 3 {
			out = append(out, strings.TrimSpace(m[1]+" "+m[2]))
		}
	}
	return out
}

func NormalizeHeadingText(heading string) string {
	parts := strings.SplitN(heading, " ", 2)
	if len(parts) < 2 {
		return heading
	}
	return strings.ToLower(strings.TrimSpace(parts[1]))
}

func CountTableColumns(body string) []int {
	var cols []int
	var currentTableLines []string
	inTable := false

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			if !inTable {
				currentTableLines = []string{}
				inTable = true
			}
			currentTableLines = append(currentTableLines, trimmed)
		} else {
			if inTable && len(currentTableLines) > 0 {
				firstRow := currentTableLines[0]
				if firstRow == "|---|" || !strings.Contains(firstRow, "---") {
					cellCount := strings.Count(firstRow, "|") - 1
					if cellCount > 0 {
						cols = append(cols, cellCount)
					}
				}
			}
			inTable = false
			currentTableLines = nil
		}
	}
	if inTable && len(currentTableLines) > 0 {
		firstRow := currentTableLines[0]
		if firstRow == "|---|" || !strings.Contains(firstRow, "---") {
			cellCount := strings.Count(firstRow, "|") - 1
			if cellCount > 0 {
				cols = append(cols, cellCount)
			}
		}
	}
	return cols
}

func CheckOmission(srcBody, tgtBody string) []Violation {
	srcWords := len(strings.Fields(srcBody))
	tgtWords := len(strings.Fields(tgtBody))
	if srcWords == 0 {
		return nil
	}
	ratio := float64(tgtWords) / float64(srcWords)
	if ratio < 0.5 {
		return []Violation{{
			Field:        "omission",
			Section:      "body",
			Message:      fmt.Sprintf("target has %d%% of source word count (%d vs %d words)", int(ratio*100), tgtWords, srcWords),
			SuggestedFix: "check for missing sections, examples, or explanatory text",
		}}
	}
	return nil
}

type Violation struct {
	Field        string
	Section      string
	Message      string
	SuggestedFix string
}

func UniqueStrings(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}
