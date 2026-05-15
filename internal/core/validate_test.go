package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loustack17/content-i18n/internal/config"
)

func TestValidateSiteConfig_HugoAdapterRequiresCanonical(t *testing.T) {
	cfg := &config.Config{
		Adapter: config.AdapterConfig{Name: "hugo"},
		URLPolicy: config.URLPolicyConfig{
			Canonical: nil,
		},
	}

	warnings := ValidateSiteConfig(cfg)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0] != "hugo adapter requires url_policy.canonical to be set" {
		t.Fatalf("unexpected warning: %s", warnings[0])
	}
}

func TestValidateSiteConfig_RuntimeTranslationEnabledButRouteEmpty(t *testing.T) {
	cfg := &config.Config{
		Adapter: config.AdapterConfig{Name: "generic"},
		URLPolicy: config.URLPolicyConfig{
			Canonical: map[string]string{"en": "/"},
			RuntimeTranslation: config.RuntimeTranslationConfig{
				Enabled:            true,
				Route:              "",
				CanonicalLanguages: []string{"en"},
			},
		},
	}

	warnings := ValidateSiteConfig(cfg)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0] != "runtime_translation enabled but route is empty" {
		t.Fatalf("unexpected warning: %s", warnings[0])
	}
}

func TestValidateSiteConfig_RuntimeTranslationEnabledButCanonicalLanguagesEmpty(t *testing.T) {
	cfg := &config.Config{
		Adapter: config.AdapterConfig{Name: "generic"},
		URLPolicy: config.URLPolicyConfig{
			Canonical: map[string]string{"en": "/"},
			RuntimeTranslation: config.RuntimeTranslationConfig{
				Enabled:            true,
				Route:              "/translate",
				CanonicalLanguages: []string{},
			},
		},
	}

	warnings := ValidateSiteConfig(cfg)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0] != "runtime_translation enabled but canonical_languages is empty" {
		t.Fatalf("unexpected warning: %s", warnings[0])
	}
}

func TestValidateSiteConfig_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		Adapter: config.AdapterConfig{Name: "hugo"},
		URLPolicy: config.URLPolicyConfig{
			Canonical: map[string]string{"en": "/", "zh-TW": "/zh-tw/"},
			RuntimeTranslation: config.RuntimeTranslationConfig{
				Enabled:            false,
				Route:              "",
				CanonicalLanguages: []string{},
			},
		},
	}

	warnings := ValidateSiteConfig(cfg)
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestValidateSite_HugoNotFound(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	os.Setenv("PATH", "/nonexistent")

	cfg := &config.Config{}
	_, err := ValidateSite(cfg, t.TempDir())
	if err == nil {
		t.Fatal("expected error when hugo not found")
	}
}

func TestValidateSite_PublicDirMissing(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		URLPolicy: config.URLPolicyConfig{
			Canonical:          map[string]string{"en": "/"},
			RuntimeTranslation: config.RuntimeTranslationConfig{},
		},
	}

	_, err := ValidateSite(cfg, tmpDir)
	if err == nil {
		t.Fatal("expected error when public dir missing")
	}
}

func setupHugoProject(t *testing.T, tmpDir string) {
	t.Helper()
	os.MkdirAll(filepath.Join(tmpDir, "content"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "hugo.toml"), []byte(`baseURL = "https://example.org/"
title = "Test Site"
`), 0644)
}

func TestValidateSite_CanonicalPathMissing(t *testing.T) {
	tmpDir := t.TempDir()
	publicDir := filepath.Join(tmpDir, "public")
	os.MkdirAll(publicDir, 0755)

	cfg := &config.Config{
		URLPolicy: config.URLPolicyConfig{
			Canonical:          map[string]string{"en": "/"},
			RuntimeTranslation: config.RuntimeTranslationConfig{},
		},
	}

	result, err := validateSitePaths(cfg, publicDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected validation to fail when canonical index.html missing")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
}

func TestValidateSite_CanonicalPathPresent(t *testing.T) {
	tmpDir := t.TempDir()
	publicDir := filepath.Join(tmpDir, "public")
	os.MkdirAll(publicDir, 0755)
	os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("test"), 0644)

	cfg := &config.Config{
		URLPolicy: config.URLPolicyConfig{
			Canonical:          map[string]string{"en": "/"},
			RuntimeTranslation: config.RuntimeTranslationConfig{},
		},
	}

	result, err := validateSitePaths(cfg, publicDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected validation to pass, got violations: %v", result.Violations)
	}
}

func TestValidateSite_UnexpectedCanonicalPath(t *testing.T) {
	tmpDir := t.TempDir()
	publicDir := filepath.Join(tmpDir, "public")
	os.MkdirAll(filepath.Join(publicDir, "en"), 0755)
	os.WriteFile(filepath.Join(publicDir, "en", "index.html"), []byte("test"), 0644)

	cfg := &config.Config{
		Project: config.ProjectConfig{
			TargetLanguages: []string{"en"},
		},
		URLPolicy: config.URLPolicyConfig{
			Canonical:          map[string]string{"zh-TW": "/zh-tw/"},
			RuntimeTranslation: config.RuntimeTranslationConfig{},
		},
	}

	result, err := validateSitePaths(cfg, publicDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected validation to fail when unexpected canonical path exists")
	}
	if len(result.Violations) < 1 {
		t.Fatalf("expected at least 1 violation, got %d", len(result.Violations))
	}
	found := false
	for _, v := range result.Violations {
		if strings.Contains(v, "unexpected canonical path") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected unexpected canonical path violation, got: %v", result.Violations)
	}
}

func TestValidateSite_RuntimeTranslationEnabledWithoutRoute(t *testing.T) {
	tmpDir := t.TempDir()
	publicDir := filepath.Join(tmpDir, "public")
	os.MkdirAll(publicDir, 0755)

	cfg := &config.Config{
		URLPolicy: config.URLPolicyConfig{
			Canonical: map[string]string{},
			RuntimeTranslation: config.RuntimeTranslationConfig{
				Enabled: true,
				Route:   "",
			},
		},
	}

	result, err := validateSitePaths(cfg, publicDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Fatal("expected validation to fail when runtime translation enabled without route")
	}
}

func TestValidateContent_ToneChecksWired(t *testing.T) {
	tmpDir := t.TempDir()

	targetPath := filepath.Join(tmpDir, "bad-tone.md")
	targetContent := `---
title: Bad Tone Test
draft: true
source_lang: zh-TW
target_lang: en
---
## Overview of the system

The system is designed to handle requests.
The user is expected to follow guidelines.
The approach is based on identity management.
The mechanism is critical for identity verification.
Identity is important for identity providers.
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Style: config.StyleConfig{
			Tone: config.ToneConfig{
				AbstractOpenerThreshold: 3,
				AbstractTerms:           []string{"identity"},
				HeadingDocLikePrefixes:  []string{"overview"},
			},
		},
	}

	result, err := ValidateContent(targetPath, &ValidateOptions{Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed {
		t.Fatalf("expected tone validation to fail, but got PASS. Violations: %v", result.Violations)
	}

	foundTone := false
	for _, v := range result.Violations {
		if v.Field == "tone" {
			foundTone = true
			break
		}
	}
	if !foundTone {
		t.Fatalf("expected tone violations, got: %v", result.Violations)
	}
}

func TestValidateContent_ToneChecksClean(t *testing.T) {
	tmpDir := t.TempDir()

	targetPath := filepath.Join(tmpDir, "clean-tone.md")
	targetContent := `---
title: Clean Tone Test
draft: true
source_lang: zh-TW
target_lang: en
---
## How we reduced latency by 40%%

We replaced the connection pool with a shared channel. The benchmark shows a 40%% improvement.
`
	if err := os.WriteFile(targetPath, []byte(targetContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Style: config.StyleConfig{
			Tone: config.ToneConfig{
				AbstractOpenerThreshold: 3,
				AbstractTerms:           []string{"identity"},
				HeadingDocLikePrefixes:  []string{"overview"},
			},
		},
	}

	result, err := ValidateContent(targetPath, &ValidateOptions{Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed {
		t.Fatalf("expected clean content to pass, got violations: %v", result.Violations)
	}
}
