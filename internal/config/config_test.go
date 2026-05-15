package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "content-i18n.yaml")
	data := `
project:
  type: generic-markdown
  source_language: zh-TW
  target_languages:
    - en
paths:
  source: .
  targets:
    en: docs/en
translation:
  default_provider: ai-harness
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Project.Type != "generic-markdown" {
		t.Errorf("type = %q, want generic-markdown", cfg.Project.Type)
	}
	if cfg.Project.SourceLanguage != "zh-TW" {
		t.Errorf("source_language = %q, want zh-TW", cfg.Project.SourceLanguage)
	}
	if len(cfg.Project.TargetLanguages) != 1 || cfg.Project.TargetLanguages[0] != "en" {
		t.Errorf("target_languages = %v, want [en]", cfg.Project.TargetLanguages)
	}
}

func TestLoad_MissingSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "content-i18n.yaml")
	data := `
project:
  type: generic-markdown
  source_language: zh-TW
  target_languages:
    - en
paths:
  source: /nonexistent
  targets:
    en: docs/en
translation:
  default_provider: ai-harness
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestLoad_MissingRequiredField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "content-i18n.yaml")
	data := `
project:
  type: generic-markdown
  source_language: zh-TW
paths:
  source: .
  targets:
    en: docs/en
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing target_languages")
	}
}
