package frontmatter

import (
	"strings"
	"testing"
)

func TestSplitNoFrontmatter(t *testing.T) {
	doc, err := Split("# Hello\n\nContent here")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Frontmatter != "" {
		t.Fatalf("expected empty frontmatter, got %q", doc.Frontmatter)
	}
	if doc.Body != "# Hello\n\nContent here" {
		t.Fatalf("unexpected body: %q", doc.Body)
	}
}

func TestSplitWithFrontmatter(t *testing.T) {
	input := "---\ntitle: Test\ndraft: true\n---\n# Hello"
	doc, err := Split(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Metadata.Title != "Test" {
		t.Fatalf("expected title=Test, got %q", doc.Metadata.Title)
	}
	if doc.Metadata.Draft != true {
		t.Fatalf("expected draft=true, got %v", doc.Metadata.Draft)
	}
	if doc.Body != "# Hello" {
		t.Fatalf("unexpected body: %q", doc.Body)
	}
}

func TestSplitMalformedFrontmatter(t *testing.T) {
	input := "---\ntitle: [invalid yaml\n---\n# Hello"
	_, err := Split(input)
	if err == nil {
		t.Fatal("expected error for malformed frontmatter")
	}
}

func TestInjectProviderMeta(t *testing.T) {
	doc, _ := Split("---\ntitle: Test\n---\n# Hello")

	pm := ProviderMeta{
		Provider: "google",
		Quality:  "machine_draft",
		Reviewed: false,
		Draft:    true,
	}

	result, err := InjectProviderMeta(doc, pm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "translation_provider: google") {
		t.Fatalf("expected translation_provider in output: %s", result)
	}
	if !strings.Contains(result, "translation_quality: machine_draft") {
		t.Fatalf("expected translation_quality in output: %s", result)
	}
	if !strings.Contains(result, "draft: true") {
		t.Fatalf("expected draft: true in output: %s", result)
	}
	if !strings.Contains(result, "reviewed: false") {
		t.Fatalf("expected reviewed: false in output: %s", result)
	}
}

func TestInjectProviderMetaPreservesBody(t *testing.T) {
	doc, _ := Split("---\ntitle: Test\n---\n# Hello\n\nContent here")

	pm := ProviderMeta{
		Provider: "deepl",
		Quality:  "machine_draft",
		Reviewed: false,
		Draft:    true,
	}

	result, err := InjectProviderMeta(doc, pm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(result, "# Hello\n\nContent here") {
		t.Fatalf("expected body preserved, got: %s", result)
	}
}

func TestInjectProviderMetaPreservesUnknownFields(t *testing.T) {
	input := `---
title: Test
date: 2026-05-15
slug: my-post
tags:
  - gcp
draft: true
---
# Hello`
	doc, _ := Split(input)

	pm := ProviderMeta{
		Provider: "google",
		Quality:  "machine_draft",
		Reviewed: false,
		Draft:    true,
	}

	result, err := InjectProviderMeta(doc, pm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "date:") {
		t.Fatalf("expected 'date:' preserved, got: %s", result)
	}
	if !strings.Contains(result, "slug:") {
		t.Fatalf("expected 'slug:' preserved, got: %s", result)
	}
	if !strings.Contains(result, "tags:") {
		t.Fatalf("expected 'tags:' preserved, got: %s", result)
	}
	if !strings.Contains(result, "gcp") {
		t.Fatalf("expected 'gcp' tag preserved, got: %s", result)
	}
	if !strings.Contains(result, "translation_provider:") {
		t.Fatalf("expected translation_provider added, got: %s", result)
	}
}
