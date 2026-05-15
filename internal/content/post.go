package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	StatusMissing  FileStatus = "missing"
	StatusStale    FileStatus = "stale"
	StatusExists   FileStatus = "exists"
	StatusInvalid  FileStatus = "invalid"
)

type FileInfo struct {
	SourcePath string
	TargetPath string
	Language   string
	Status     FileStatus
}

func Discover(cfg *config.Config) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(cfg.Paths.Source, func(path string, d fs.DirEntry, err error) error {
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

		for _, lang := range cfg.Project.TargetLanguages {
			targetDir, ok := cfg.Paths.Targets[lang]
			if !ok {
				continue
			}

			targetPath := filepath.Join(targetDir, rel)
			if cfg.Adapter.PreserveRelativePaths {
				targetPath = filepath.Join(targetDir, rel)
			}

			status := computeStatus(path, targetPath)
			files = append(files, FileInfo{
				SourcePath: path,
				TargetPath: targetPath,
				Language:   lang,
				Status:     status,
			})
		}

		return nil
	})

	return files, err
}

func computeStatus(sourcePath, targetPath string) FileStatus {
	srcInfo, err := os.Stat(sourcePath)
	if err != nil {
		return StatusInvalid
	}

	tgtInfo, err := os.Stat(targetPath)
	if err != nil {
		return StatusMissing
	}

	if srcInfo.ModTime().After(tgtInfo.ModTime()) {
		return StatusStale
	}

	return StatusExists
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
	ProjectType      string
	SourceLanguage   string
	SourcePath       string
	TargetLanguages  []string
	TargetPaths      map[string]string
	SourceFileCount  int
	TargetFileCount  int
	MissingCount     int
}

func BuildStatusReport(cfg *config.Config, files []FileInfo) *StatusReport {
	counts := CountByStatus(files)
	return &StatusReport{
		ProjectType:     cfg.Project.Type,
		SourceLanguage:  cfg.Project.SourceLanguage,
		SourcePath:      cfg.Paths.Source,
		TargetLanguages: cfg.Project.TargetLanguages,
		TargetPaths:     cfg.Paths.Targets,
		SourceFileCount: counts[StatusExists] + counts[StatusStale] + counts[StatusMissing] + counts[StatusInvalid],
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

func FileMtime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
