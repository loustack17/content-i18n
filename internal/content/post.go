package content

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
)

type Post struct {
	Path        string
	Language    string
	Frontmatter string
	Body        string
}

type FileStatus string

const (
	StatusMissing FileStatus = "missing"
	StatusStale   FileStatus = "stale"
	StatusExists  FileStatus = "exists"
	StatusInvalid FileStatus = "invalid"
)

type FileInfo struct {
	SourcePath string
	TargetPath string
	Language   string
	Status     FileStatus
	SourceHash string
}

type statusStore struct {
	Entries map[string]string `json:"entries"`
}

func loadStatusStore(path string) (*statusStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &statusStore{Entries: make(map[string]string)}, nil
		}
		return nil, err
	}
	var store statusStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("corrupt status file: %w", err)
	}
	if store.Entries == nil {
		store.Entries = make(map[string]string)
	}
	return &store, nil
}

func statusKey(relPath, lang string) string {
	return relPath + ":" + lang
}

func Discover(cfg *config.Config) ([]FileInfo, error) {
	var files []FileInfo

	store, err := loadStatusStore(cfg.StatusFilePath())
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(cfg.Paths.Source, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		rel, err := filepath.Rel(cfg.Paths.Source, path)
		if err != nil {
			return err
		}

		srcHash, err := FileHash(path)
		if err != nil {
			return fmt.Errorf("hash %s: %w", path, err)
		}

		for _, lang := range cfg.Project.TargetLanguages {
			targetPath, err := TargetPath(cfg, path, lang)
			if err != nil {
				return err
			}

			status := computeStatus(targetPath, lang, srcHash, rel, store)
			files = append(files, FileInfo{
				SourcePath: path,
				TargetPath: targetPath,
				Language:   lang,
				Status:     status,
				SourceHash: srcHash,
			})
		}

		return nil
	})

	return files, err
}

func computeStatus(targetPath, lang, srcHash, rel string, store *statusStore) FileStatus {
	_, err := os.Stat(targetPath)
	if err != nil {
		return StatusMissing
	}

	key := statusKey(rel, lang)
	storedHash, ok := store.Entries[key]
	if !ok {
		return StatusStale
	}

	if srcHash != storedHash {
		return StatusStale
	}

	return StatusExists
}

func FileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func WriteStatusEntry(cfg *config.Config, sourcePath, lang, sourceHash string) error {
	rel, err := filepath.Rel(cfg.Paths.Source, sourcePath)
	if err != nil {
		return err
	}
	key := statusKey(rel, lang)

	store, err := loadStatusStore(cfg.StatusFilePath())
	if err != nil {
		return err
	}
	store.Entries[key] = sourceHash

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	statusDir := filepath.Dir(cfg.StatusFilePath())
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(cfg.StatusFilePath(), data, 0644)
}

func CountByStatus(files []FileInfo) map[FileStatus]int {
	counts := make(map[FileStatus]int)
	for _, f := range files {
		counts[f.Status]++
	}
	return counts
}

func MissingTranslations(files []FileInfo) []FileInfo {
	var missing []FileInfo
	for _, f := range files {
		if f.Status == StatusMissing || f.Status == StatusStale {
			missing = append(missing, f)
		}
	}
	return missing
}

type StatusReport struct {
	ProjectType     string
	SourceLanguage  string
	SourcePath      string
	TargetLanguages []string
	TargetPaths     map[string]string
	SourceFileCount int
	TargetFileCount int
	MissingCount    int
}

func countUniqueSources(files []FileInfo) int {
	seen := make(map[string]struct{})
	for _, f := range files {
		seen[f.SourcePath] = struct{}{}
	}
	return len(seen)
}

func BuildStatusReport(cfg *config.Config, files []FileInfo) *StatusReport {
	counts := CountByStatus(files)
	return &StatusReport{
		ProjectType:     cfg.Project.Type,
		SourceLanguage:  cfg.Project.SourceLanguage,
		SourcePath:      cfg.Paths.Source,
		TargetLanguages: cfg.Project.TargetLanguages,
		TargetPaths:     cfg.Paths.Targets,
		SourceFileCount: countUniqueSources(files),
		TargetFileCount: counts[StatusExists] + counts[StatusStale],
		MissingCount:    counts[StatusMissing] + counts[StatusStale],
	}
}

func TargetPath(cfg *config.Config, sourcePath, lang string) (string, error) {
	targetDir, ok := cfg.Paths.Targets[lang]
	if !ok {
		return "", fmt.Errorf("no target path for language %s", lang)
	}

	rel, err := filepath.Rel(cfg.Paths.Source, sourcePath)
	if err != nil {
		return "", err
	}

	return filepath.Join(targetDir, rel), nil
}
