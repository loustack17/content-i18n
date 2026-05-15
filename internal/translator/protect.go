package translator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type TokenMap struct {
	tokens map[string]string
}

func Protect(text string) (string, *TokenMap) {
	tm := &TokenMap{tokens: make(map[string]string)}
	result := text

	result = tm.replaceFencedCodeBlocks(result)
	result = tm.replaceInlineCode(result)
	result = tm.replaceURLs(result)

	return result, tm
}

func Restore(text string, tm *TokenMap) (string, error) {
	result := text
	used := make(map[string]bool)

	keys := make([]string, 0, len(tm.tokens))
	for k := range tm.tokens {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	for _, placeholder := range keys {
		if !strings.Contains(result, placeholder) {
			return "", fmt.Errorf("placeholder %s not found in text", placeholder)
		}
		count := strings.Count(result, placeholder)
		if count > 1 {
			return "", fmt.Errorf("placeholder %s appears %d times, expected 1", placeholder, count)
		}
		result = strings.Replace(result, placeholder, tm.tokens[placeholder], 1)
		used[placeholder] = true
	}

	return result, nil
}

func (tm *TokenMap) Get(placeholder string) (string, bool) {
	v, ok := tm.tokens[placeholder]
	return v, ok
}

func (tm *TokenMap) add(prefix string, original string) string {
	idx := len(tm.tokens)
	placeholder := fmt.Sprintf("__%s_%d__", prefix, idx)
	tm.tokens[placeholder] = original
	return placeholder
}

var fencePattern = regexp.MustCompile("(?s)```.*?```")

func (tm *TokenMap) replaceFencedCodeBlocks(text string) string {
	return fencePattern.ReplaceAllStringFunc(text, func(match string) string {
		return tm.add("CODE_BLOCK", match)
	})
}

var inlineCodePattern = regexp.MustCompile("`([^`\n]*)`")

func (tm *TokenMap) replaceInlineCode(text string) string {
	return inlineCodePattern.ReplaceAllStringFunc(text, func(match string) string {
		return tm.add("INLINE_CODE", match)
	})
}

var urlPattern = regexp.MustCompile(`https?://[^\s\)\]"<>]+`)

func (tm *TokenMap) replaceURLs(text string) string {
	return urlPattern.ReplaceAllStringFunc(text, func(match string) string {
		trailing := strings.TrimRight(match, ".,;:!?")
		suffix := match[len(trailing):]
		placeholder := tm.add("URL", trailing)
		return placeholder + suffix
	})
}
