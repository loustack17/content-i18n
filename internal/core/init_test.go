package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/loustack17/content-i18n/internal/config"
)

func TestInit_GenericMarkdown(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "content-i18n.yaml")
	os.MkdirAll(filepath.Join(dir, "docs", "zh-TW"), 0755)
	os.MkdirAll(filepath.Join(dir, "docs", "en"), 0755)

	result, err := Init(InitOptions{
		Type:   "generic-markdown",
		Output: out,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Created) == 0 {
		t.Fatal("expected files created")
	}

	cfg, err := config.Load(out)
	if err != nil {
		t.Fatalf("generated config failed to load: %v", err)
	}
	if cfg.Project.Type != "generic-markdown" {
		t.Fatalf("expected generic-markdown type, got %s", cfg.Project.Type)
	}
}

func TestInit_Hugo(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "content-i18n.yaml")
	os.MkdirAll(filepath.Join(dir, "content", "zh-TW"), 0755)
	os.MkdirAll(filepath.Join(dir, "content", "en"), 0755)

	result, err := Init(InitOptions{
		Type:   "hugo",
		Output: out,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Created) == 0 {
		t.Fatal("expected files created")
	}

	cfg, err := config.Load(out)
	if err != nil {
		t.Fatalf("generated config failed to load: %v", err)
	}
	if cfg.Project.Type != "hugo" {
		t.Fatalf("expected hugo type, got %s", cfg.Project.Type)
	}
	if cfg.Adapter.Name != "hugo" {
		t.Fatalf("expected hugo adapter, got %s", cfg.Adapter.Name)
	}
}

func TestInit_ProtectsExistingFiles(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "content-i18n.yaml")

	_, err := Init(InitOptions{
		Type:   "generic-markdown",
		Output: out,
	})
	if err != nil {
		t.Fatal(err)
	}

	orig, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	result, err := Init(InitOptions{
		Type:   "hugo",
		Output: out,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Skipped) == 0 {
		t.Fatal("expected files skipped")
	}

	current, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != string(orig) {
		t.Fatal("existing file was overwritten")
	}
}

func TestInit_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "content-i18n.yaml")
	os.MkdirAll(filepath.Join(dir, "docs", "zh-TW"), 0755)
	os.MkdirAll(filepath.Join(dir, "docs", "en"), 0755)
	os.MkdirAll(filepath.Join(dir, "content", "zh-TW"), 0755)
	os.MkdirAll(filepath.Join(dir, "content", "en"), 0755)

	_, err := Init(InitOptions{
		Type:   "generic-markdown",
		Output: out,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := Init(InitOptions{
		Type:   "hugo",
		Output: out,
		Force:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Skipped) > 0 {
		t.Fatalf("expected no skipped files with force, got %v", result.Skipped)
	}

	cfg, err := config.Load(out)
	if err != nil {
		t.Fatalf("generated config failed to load: %v", err)
	}
	if cfg.Project.Type != "hugo" {
		t.Fatalf("expected hugo type after force, got %s", cfg.Project.Type)
	}
}

func TestInit_UnknownType(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "content-i18n.yaml")

	_, err := Init(InitOptions{
		Type:   "unknown",
		Output: out,
	})
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestInit_CreatesSupportFiles(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "content-i18n.yaml")

	result, err := Init(InitOptions{
		Type:   "generic-markdown",
		Output: out,
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		".content-i18n/glossary.yaml",
		".content-i18n/style/technical-english.yaml",
	}

	for _, rel := range expected {
		full := filepath.Join(dir, rel)
		if _, err := os.Stat(full); err != nil {
			t.Fatalf("expected support file %s: %v", rel, err)
		}
	}

	createdSet := make(map[string]bool)
	for _, c := range result.Created {
		createdSet[c] = true
	}
	for _, rel := range expected {
		full := filepath.Join(dir, rel)
		if !createdSet[full] {
			t.Fatalf("expected %s in created files", full)
		}
	}
}
