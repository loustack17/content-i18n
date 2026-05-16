package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
)

type QueueEntry struct {
	SourcePath string             `json:"source_path"`
	TargetPath string             `json:"target_path"`
	Language   string             `json:"language"`
	Status     content.FileStatus `json:"status"`
	SourceHash string             `json:"source_hash"`
}

type BatchStatus struct {
	Total     int          `json:"total"`
	Completed int          `json:"completed"`
	Stale     int          `json:"stale"`
	Missing   int          `json:"missing"`
	Next      *QueueEntry  `json:"next"`
	Queue     []QueueEntry `json:"queue"`
}

func TranslationQueue(cfg *config.Config, group string) (*BatchStatus, error) {
	files, err := content.Discover(cfg)
	if err != nil {
		return nil, err
	}

	var filtered []content.FileInfo
	for _, f := range files {
		if group != "" && !matchesGroup(f.SourcePath, group) {
			continue
		}
		filtered = append(filtered, f)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].SourcePath < filtered[j].SourcePath
	})

	total := len(filtered)
	completed := 0
	stale := 0
	missing := 0
	var queue []QueueEntry

	for _, f := range filtered {
		entry := QueueEntry{
			SourcePath: f.SourcePath,
			TargetPath: f.TargetPath,
			Language:   f.Language,
			Status:     f.Status,
			SourceHash: f.SourceHash,
		}

		switch f.Status {
		case content.StatusExists:
			completed++
		case content.StatusStale:
			stale++
			queue = append(queue, entry)
		case content.StatusMissing:
			missing++
			queue = append(queue, entry)
		case content.StatusInvalid:
			queue = append(queue, entry)
		}
	}

	var next *QueueEntry
	if len(queue) > 0 {
		next = &queue[0]
	}

	return &BatchStatus{
		Total:     total,
		Completed: completed,
		Stale:     stale,
		Missing:   missing,
		Next:      next,
		Queue:     queue,
	}, nil
}

func NextTranslation(cfg *config.Config, group string) (*QueueEntry, error) {
	status, err := TranslationQueue(cfg, group)
	if err != nil {
		return nil, err
	}
	return status.Next, nil
}

func matchesGroup(sourcePath string, group string) bool {
	groupLower := strings.ToLower(group)
	sourceLower := strings.ToLower(sourcePath)
	dir := strings.ToLower(filepath.Base(filepath.Dir(sourcePath)))
	parentDir := strings.ToLower(filepath.Base(filepath.Dir(filepath.Dir(sourcePath))))
	return strings.Contains(dir, groupLower) ||
		strings.Contains(parentDir, groupLower) ||
		strings.Contains(sourceLower, groupLower)
}

type QueueStore struct {
	Completed map[string]string `json:"completed"`
}

func queueStorePath(cfg *config.Config) string {
	return filepath.Join(cfg.ConfigDir, ".content-i18n", "queue.json")
}

func loadQueueStore(cfg *config.Config) (*QueueStore, error) {
	path := queueStorePath(cfg)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &QueueStore{Completed: make(map[string]string)}, nil
		}
		return nil, err
	}
	var store QueueStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("corrupt queue file: %w", err)
	}
	if store.Completed == nil {
		store.Completed = make(map[string]string)
	}
	return &store, nil
}

func saveQueueStore(cfg *config.Config, store *QueueStore) error {
	path := queueStorePath(cfg)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func MarkTranslationComplete(cfg *config.Config, sourcePath string, language string, sourceHash string) error {
	store, err := loadQueueStore(cfg)
	if err != nil {
		return err
	}
	key := sourcePath + ":" + language
	store.Completed[key] = sourceHash
	return saveQueueStore(cfg, store)
}
