package google

import (
	"os"
	"testing"
)

func TestNewMissingCredentials(t *testing.T) {
	orig := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	defer os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", orig)
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")

	_, err := New()
	if err == nil {
		t.Fatal("expected error when GOOGLE_APPLICATION_CREDENTIALS not set")
	}
}

func TestNewMissingProject(t *testing.T) {
	origCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	origProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	defer func() {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", origCreds)
		os.Setenv("GOOGLE_CLOUD_PROJECT", origProject)
	}()
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/fake/path.json")
	os.Setenv("GOOGLE_CLOUD_PROJECT", "")

	_, err := New()
	if err == nil {
		t.Fatal("expected error when GOOGLE_CLOUD_PROJECT not set")
	}
}

func TestGoogleLangCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"zh-TW", "zh-TW"},
		{"zh-tw", "zh-TW"},
		{"zh-CN", "zh-CN"},
		{"en", "en"},
		{"ja", "ja"},
		{"ko", "ko"},
		{"de", "de"},
		{"fr", "fr"},
		{"es", "es"},
		{"pt-BR", "pt-BR"},
		{"pt-br", "pt-BR"},
		{"pt", "pt"},
		{"ru", "ru"},
		{"it", "it"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		got := googleLangCode(tt.input)
		if got != tt.want {
			t.Errorf("googleLangCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTranslateBatchEmpty(t *testing.T) {
	p := &Provider{}
	result, err := p.TranslateBatch([]string{}, "en", "zh-TW")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestCompileGlossary(t *testing.T) {
	entries := []GlossaryEntry{
		{Source: "服務帳號", Target: "service account"},
		{Source: "雲端", Target: "cloud"},
	}

	result := CompileGlossary(entries)
	expected := "服務帳號\tservice account\n雲端\tcloud"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestCompileGlossaryWithComma(t *testing.T) {
	entries := []GlossaryEntry{
		{Source: "foo,bar", Target: "service, account"},
	}

	result := CompileGlossary(entries)
	expected := "foo,bar\tservice, account"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestCompileGlossaryEmpty(t *testing.T) {
	result := CompileGlossary([]GlossaryEntry{})
	if result != "" {
		t.Fatalf("expected empty string, got %q", result)
	}
}

func TestDefaultMetadata(t *testing.T) {
	meta := DefaultMetadata("zh-TW", "en", "test-glossary")

	if meta.Provider != "google" {
		t.Fatalf("expected provider=google, got %s", meta.Provider)
	}
	if meta.Quality != "machine_draft" {
		t.Fatalf("expected quality=machine_draft, got %s", meta.Quality)
	}
	if meta.Reviewed != false {
		t.Fatalf("expected reviewed=false, got %v", meta.Reviewed)
	}
	if meta.Draft != true {
		t.Fatalf("expected draft=true, got %v", meta.Draft)
	}
	if meta.SourceLang != "zh-TW" {
		t.Fatalf("expected source_lang=zh-TW, got %s", meta.SourceLang)
	}
	if meta.TargetLang != "en" {
		t.Fatalf("expected target_lang=en, got %s", meta.TargetLang)
	}
	if meta.GlossaryID != "test-glossary" {
		t.Fatalf("expected glossary_id=test-glossary, got %s", meta.GlossaryID)
	}
}
