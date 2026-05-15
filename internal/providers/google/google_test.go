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
