package deepl

import (
	"os"
	"testing"
)

func TestNewMissingAPIKey(t *testing.T) {
	orig := os.Getenv("DEEPL_API_KEY")
	defer os.Setenv("DEEPL_API_KEY", orig)
	os.Setenv("DEEPL_API_KEY", "")

	_, err := New()
	if err == nil {
		t.Fatal("expected error when DEEPL_API_KEY not set")
	}
}

func TestNewWithAPIKey(t *testing.T) {
	orig := os.Getenv("DEEPL_API_KEY")
	defer os.Setenv("DEEPL_API_KEY", orig)
	os.Setenv("DEEPL_API_KEY", "test-key")

	p, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.Available() {
		t.Fatal("expected provider to be available")
	}
}

func TestCompileGlossary(t *testing.T) {
	entries := []GlossaryEntry{
		{Source: "服務帳號", Target: "service account"},
		{Source: "雲端", Target: "cloud"},
	}

	result := CompileGlossary(entries)
	expected := "服務帳號,service account\n雲端,cloud"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestCompileGlossaryWithComma(t *testing.T) {
	entries := []GlossaryEntry{
		{Source: "foo,bar", Target: "service, account"},
	}

	result := CompileGlossary(entries)
	expected := `"foo,bar","service, account"`
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestCompileGlossaryWithQuote(t *testing.T) {
	entries := []GlossaryEntry{
		{Source: `foo"bar`, Target: `he said "hi"`},
	}

	result := CompileGlossary(entries)
	expected := `"foo""bar","he said ""hi"""`
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

func TestDeeplLangCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"en", "EN-US"},
		{"EN", "EN-US"},
		{"zh-TW", "ZH"},
		{"ZH-TW", "ZH"},
		{"zh-CN", "ZH"},
		{"ja", "JA"},
		{"de", "DE"},
		{"fr", "FR"},
		{"pt-BR", "PT-BR"},
		{"pt", "PT-PT"},
		{"unknown", "UNKNOWN"},
	}

	for _, tt := range tests {
		got := deeplLangCode(tt.input)
		if got != tt.want {
			t.Errorf("deeplLangCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTranslateBatchEmpty(t *testing.T) {
	orig := os.Getenv("DEEPL_API_KEY")
	defer os.Setenv("DEEPL_API_KEY", orig)
	os.Setenv("DEEPL_API_KEY", "test-key")

	p, _ := New()
	result, err := p.TranslateBatch([]string{}, "en", "zh-TW")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}
