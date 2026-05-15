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
