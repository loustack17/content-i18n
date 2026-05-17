package core

import (
	"fmt"
	"os"
	"path/filepath"
)

const hugoConfig = `project:
  type: hugo
  source_language: zh-TW
  target_languages:
    - en

paths:
  source: content/zh-TW
  targets:
    en: content/en

adapter:
  name: hugo
  mode: content_directory
  preserve_relative_paths: true
  translation_key: contentbasename

url_policy:
  canonical:
    en: /
    zh-TW: /zh_tw/

translation:
  default_provider: ai-harness
  fallback_providers:
    - deepl
    - google
  output:
    draft: true
    reviewed: false
    preserve_code_blocks: true
    preserve_inline_code: true
    preserve_frontmatter_keys: true
    preserve_links: true

style:
  pack: .content-i18n/style/technical-english.yaml
  glossary: .content-i18n/glossary.yaml
  tone:
    abstract_opener_threshold: 3
    abstract_terms:
      - identity
      - solution
      - approach
      - mechanism
    heading_doc_like_prefixes:
      - overview of
      - introduction to
      - prerequisites for
      - what is
`

const genericMarkdownConfig = `project:
  type: generic-markdown
  source_language: zh-TW
  target_languages:
    - en

paths:
  source: docs/zh-TW
  targets:
    en: docs/en

adapter:
  name: generic-markdown
  preserve_relative_paths: true

translation:
  default_provider: ai-harness
  fallback_providers:
    - deepl
    - google
  output:
    draft: true
    reviewed: false
    preserve_code_blocks: true
    preserve_inline_code: true
    preserve_links: true

style:
  pack: .content-i18n/style/technical-english.yaml
  glossary: .content-i18n/glossary.yaml
  tone:
    abstract_opener_threshold: 3
    abstract_terms:
      - identity
      - solution
      - approach
      - mechanism
    heading_doc_like_prefixes:
      - overview of
      - introduction to
      - what is
`

const glossaryYAML = `terms:
  - source: "部署"
    target: "deployment"
    source_lang: zh-TW
    target_lang: en
    note: "Use deployment as noun. Use deploy only as verb."
  - source: "維運"
    target: "operations"
    source_lang: zh-TW
    target_lang: en
  - source: "可靠性"
    target: "reliability"
    source_lang: zh-TW
    target_lang: en
  - source: "可觀測性"
    target: "observability"
    source_lang: zh-TW
    target_lang: en
`

const stylePackYAML = `name: technical-english
voice:
  audience: "Technical readers, developers, and engineers"
  tone: "clear, direct, technically accurate"
  rewrite_level: "translate faithfully into natural target-language prose without changing structure, content coverage, argument flow, emphasis, or style class"
prefer_terms: []
avoid:
  - unexplained acronyms
  - literal source-language sentence order
  - marketing language
rules:
  titles: "Translate titles into natural technical English. Keep the original meaning."
  headings: "Keep heading hierarchy exactly. Translate heading text. Do not add, remove, merge, or split heading levels."
  code: "Do not translate code, commands, config keys, package names, API names, resource names, or error strings."
  links: "Preserve URLs exactly. Translate link text without changing meaning or emphasis."
  frontmatter: "Translate title, description, summary, keywords. Preserve date, slug, aliases, tags as-is."
  structure: "Preserve paragraph count, list count and nesting, table dimensions, blockquotes, and horizontal rules."
  examples: "Keep all examples and code samples in the same order. Translate surrounding explanatory text without changing technical meaning."
  argument_flow: "Keep the source's section order and reasoning flow. Do not reorder, merge, or split sections."
  style_class: "Match the source's genre. Do not add editorial commentary, opinions, or summaries absent from the source."
`

type InitOptions struct {
	Type   string
	Output string
	Force  bool
}

type InitResult struct {
	Created []string
	Skipped []string
}

func Init(opts InitOptions) (*InitResult, error) {
	var result InitResult

	switch opts.Type {
	case "hugo":
	case "generic-markdown":
	default:
		return nil, fmt.Errorf("unknown type %q (use hugo or generic-markdown)", opts.Type)
	}

	out := opts.Output
	if out == "" {
		out = "content-i18n.yaml"
	}

	configContent := genericMarkdownConfig
	if opts.Type == "hugo" {
		configContent = hugoConfig
	}

	if err := writeIfAbsent(out, []byte(configContent), opts.Force, &result); err != nil {
		return &result, err
	}

	supportFiles := map[string]string{
		".content-i18n/glossary.yaml":                glossaryYAML,
		".content-i18n/style/technical-english.yaml": stylePackYAML,
	}

	for relPath, content := range supportFiles {
		target := filepath.Join(filepath.Dir(out), relPath)
		if err := writeIfAbsent(target, []byte(content), opts.Force, &result); err != nil {
			return &result, err
		}
	}

	return &result, nil
}

func writeIfAbsent(path string, data []byte, force bool, result *InitResult) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			result.Skipped = append(result.Skipped, path)
			return nil
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	result.Created = append(result.Created, path)
	return nil
}
