package content

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/loustack/content-i18n/internal/config"
)

func TestDiscover(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	tgtDir := filepath.Join(dir, "tgt")
	os.MkdirAll(filepath.Join(srcDir, "posts"), 0755)
	os.MkdirAll(tgtDir, 0755)

	os.WriteFile(filepath.Join(srcDir, "a.md"), []byte("# A\n"), 0644)
	os.WriteFile(filepath.Join(srcDir, "posts", "b.md"), []byte("# B\n"), 0644)

	cfg := &config.Config{
		Project: config.ProjectConfig{
			Type:            "generic-markdown",
			SourceLanguage:  "zh-TW",
			TargetLanguages: []string{"en"},
		},
		Paths: config.PathsConfig{
			Source:  srcDir,
			Targets: map[string]string{"en": tgtDir},
		},
		Adapter: config.AdapterConfig{PreserveRelativePaths: true},
		Translation: config.TranslationConfig{
			DefaultProvider: "ai-harness",
		},
	}

	files, err := Discover(cfg)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	counts := CountByStatus(files)
	if counts[StatusMissing] != 2 {
		t.Errorf("expected 2 missing, got %d", counts[StatusMissing])
	}
}

func TestTargetPath(t *testing.T) {
	cfg := &config.Config{
		Paths: config.PathsConfig{
			Source:  "/content/zh-TW",
			Targets: map[string]string{"en": "/content/en"},
		},
	}
	path, err := TargetPath(cfg, "/content/zh-TW/posts/hello.md", "en")
	if err != nil {
		t.Fatalf("target path: %v", err)
	}
	want := filepath.Join("/content/en", "posts", "hello.md")
	if path != want {
		t.Errorf("target path = %q, want %q", path, want)
	}
}

func TestComputeStatus(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.md")
	tgt := filepath.Join(dir, "tgt.md")

	os.WriteFile(src, []byte("hello"), 0644)
	os.WriteFile(tgt, []byte("hello"), 0644)

	past := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(tgt, past, past); err != nil {
		t.Fatal(err)
	}

	status := computeStatus(src, tgt)
	if status != StatusStale {
		t.Errorf("expected stale, got %s", status)
	}
}

func TestMissingTranslations(t *testing.T) {
	files := []FileInfo{
		{Status: StatusExists},
		{Status: StatusMissing},
		{Status: StatusStale},
	}
	missing := MissingTranslations(files)
	if len(missing) != 2 {
		t.Errorf("expected 2 missing, got %d", len(missing))
	}
}