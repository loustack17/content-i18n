package content

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/loustack17/content-i18n/internal/config"
)

func testCfg(t *testing.T, srcDir, tgtDir string) *config.Config {
	t.Helper()
	return &config.Config{
		ConfigDir: t.TempDir(),
		Project: config.ProjectConfig{
			Type:            "generic-markdown",
			SourceLanguage:  "zh-TW",
			TargetLanguages: []string{"en"},
		},
		Paths: config.PathsConfig{
			Source:  srcDir,
			Targets: map[string]string{"en": tgtDir},
		},
		Translation: config.TranslationConfig{
			DefaultProvider: "ai-harness",
		},
	}
}

func TestDiscover(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	tgtDir := filepath.Join(dir, "tgt")
	os.MkdirAll(filepath.Join(srcDir, "posts"), 0755)
	os.MkdirAll(tgtDir, 0755)

	os.WriteFile(filepath.Join(srcDir, "a.md"), []byte("# A\n"), 0644)
	os.WriteFile(filepath.Join(srcDir, "posts", "b.md"), []byte("# B\n"), 0644)

	cfg := testCfg(t, srcDir, tgtDir)

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

func TestComputeStatus_HashStale(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	tgtDir := filepath.Join(dir, "tgt")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(tgtDir, 0755)

	src := filepath.Join(srcDir, "src.md")
	tgt := filepath.Join(tgtDir, "src.md")

	os.WriteFile(src, []byte("source content v1"), 0644)
	os.WriteFile(tgt, []byte("translation"), 0644)

	cfg := testCfg(t, srcDir, tgtDir)
	WriteStatusEntry(cfg, src, "en", "oldhash")

	store, err := loadStatusStore(cfg.StatusFilePath())
	if err != nil {
		t.Fatalf("load status store: %v", err)
	}
	srcHash, err := FileHash(src)
	if err != nil {
		t.Fatalf("fileHash: %v", err)
	}

	status := computeStatus(tgt, "en", srcHash, "src.md", store)
	if status != StatusStale {
		t.Errorf("expected stale for hash mismatch, got %s", status)
	}
}

func TestComputeStatus_HashMatches(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "src")
	tgtDir := filepath.Join(dir, "tgt")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(tgtDir, 0755)

	src := filepath.Join(srcDir, "src.md")
	tgt := filepath.Join(tgtDir, "src.md")

	srcContent := []byte("same content here")
	os.WriteFile(src, srcContent, 0644)
	os.WriteFile(tgt, []byte("translation"), 0644)

	cfg := testCfg(t, srcDir, tgtDir)
	srcHash, err := FileHash(src)
	if err != nil {
		t.Fatalf("fileHash: %v", err)
	}
	WriteStatusEntry(cfg, src, "en", srcHash)

	store, err := loadStatusStore(cfg.StatusFilePath())
	if err != nil {
		t.Fatalf("load status store: %v", err)
	}

	status := computeStatus(tgt, "en", srcHash, "src.md", store)
	if status != StatusExists {
		t.Errorf("expected exists for matching source hash, got %s", status)
	}
}

func TestComputeStatus_Missing(t *testing.T) {
	tgt := filepath.Join(t.TempDir(), "nonexistent", "src.md")
	store := &statusStore{Entries: make(map[string]string)}

	status := computeStatus(tgt, "en", "anyhash", "src.md", store)
	if status != StatusMissing {
		t.Errorf("expected missing, got %s", status)
	}
}

func TestComputeStatus_NoSidecar(t *testing.T) {
	dir := t.TempDir()
	tgtDir := filepath.Join(dir, "tgt")
	os.MkdirAll(tgtDir, 0755)

	tgt := filepath.Join(tgtDir, "src.md")
	os.WriteFile(tgt, []byte("translation"), 0644)

	store := &statusStore{Entries: make(map[string]string)}

	status := computeStatus(tgt, "en", "anyhash", "src.md", store)
	if status != StatusStale {
		t.Errorf("expected stale when no sidecar, got %s", status)
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

func TestBuildStatusReport_SourceFileCount(t *testing.T) {
	cfg := &config.Config{
		Project: config.ProjectConfig{
			Type:            "generic-markdown",
			SourceLanguage:  "zh-TW",
			TargetLanguages: []string{"en", "ja"},
		},
		Paths: config.PathsConfig{
			Source:  "/src",
			Targets: map[string]string{"en": "/en", "ja": "/ja"},
		},
	}
	files := []FileInfo{
		{SourcePath: "/src/a.md", Language: "en", Status: StatusMissing},
		{SourcePath: "/src/a.md", Language: "ja", Status: StatusMissing},
	}
	report := BuildStatusReport(cfg, files)
	if report.SourceFileCount != 1 {
		t.Errorf("SourceFileCount = %d, want 1", report.SourceFileCount)
	}
	if report.MissingCount != 2 {
		t.Errorf("MissingCount = %d, want 2", report.MissingCount)
	}
}
