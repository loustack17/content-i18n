package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/content"
	"github.com/loustack17/content-i18n/internal/frontmatter"
	"github.com/loustack17/content-i18n/internal/validator"
)

func ApplyWork(cfg *config.Config, slug string, dryRun bool, force bool) error {
	workDir := filepath.Join("work", slug)
	metaPath := filepath.Join(workDir, "meta.json")

	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return fmt.Errorf("read meta: %w", err)
	}

	var meta WorkMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return fmt.Errorf("parse meta: %w", err)
	}

	targetPath := filepath.Join(workDir, "target.md")
	targetData, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("read target: %w", err)
	}

	if len(strings.TrimSpace(string(targetData))) == 0 {
		return fmt.Errorf("target.md is empty, nothing to apply")
	}

	// Validate target before writing
	violations, err := validator.Validate(targetPath, meta.SourcePath, nil)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	if len(violations) > 0 && !force {
		return fmt.Errorf("validation failed: %v", violations)
	}
	if len(violations) > 0 && force {
		fmt.Printf("warning: validation issues, overriding with --force: %v\n", violations)
	}

	// Compute the actual target path under the configured language directory
	actualTarget, err := content.TargetPath(cfg, meta.SourcePath, meta.TargetLanguage)
	if err != nil {
		return fmt.Errorf("compute target path: %w", err)
	}

	if dryRun {
		fmt.Printf("would write: %s\n", actualTarget)
		existing, _ := os.ReadFile(actualTarget)
		if len(existing) > 0 {
			fmt.Printf("--- existing\n+++ new\n")
			diffLines(existing, targetData)
		} else {
			fmt.Printf("(new file, %d bytes)\n", len(targetData))
		}
		return nil
	}

	output := targetData
	if meta.Provider != "" && meta.Provider != "manual" {
		doc, err := frontmatter.Split(string(targetData))
		if err != nil {
			return fmt.Errorf("parse target frontmatter: %w", err)
		}
		injected, err := frontmatter.InjectProviderMeta(doc, frontmatter.ProviderMeta{
			Provider: meta.Provider,
			Quality:  "machine_draft",
			Reviewed: false,
			Draft:    true,
		})
		if err != nil {
			return fmt.Errorf("inject provider metadata: %w", err)
		}
		output = []byte(injected)
	}

	if err := os.MkdirAll(filepath.Dir(actualTarget), 0755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	if err := os.WriteFile(actualTarget, output, 0644); err != nil {
		return fmt.Errorf("write target: %w", err)
	}

	// Update status hash
	hash, err := content.FileHash(meta.SourcePath)
	if err != nil {
		return fmt.Errorf("compute source hash: %w", err)
	}
	if err := content.WriteStatusEntry(cfg, meta.SourcePath, meta.TargetLanguage, hash); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	fmt.Printf("wrote: %s\n", actualTarget)
	return nil
}

func diffLines(a, b []byte) {
	linesA := strings.Split(strings.TrimSuffix(string(a), "\n"), "\n")
	linesB := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")

	contextSize := 3
	type change struct {
		kind string // "equal", "delete", "add"
		line string
	}

	var changes []change
	i, j := 0, 0
	for i < len(linesA) || j < len(linesB) {
		if i < len(linesA) && j < len(linesB) && linesA[i] == linesB[j] {
			changes = append(changes, change{"equal", linesA[i]})
			i++
			j++
		} else {
			if j < len(linesB) {
				changes = append(changes, change{"add", linesB[j]})
				j++
			}
			if i < len(linesA) {
				changes = append(changes, change{"delete", linesA[i]})
				i++
			}
		}
	}

	var hunks []struct {
		oldStart, newStart int
		lines              []change
	}
	var currentLines []change
	var inHunk bool

	for idx, c := range changes {
		if c.kind != "equal" {
			if !inHunk {
				start := max(0, idx-contextSize)
				var oldStart, newStart int
				for k := 0; k < start; k++ {
					if changes[k].kind != "add" {
						oldStart++
					}
					if changes[k].kind != "delete" {
						newStart++
					}
				}
				hunks = append(hunks, struct {
					oldStart, newStart int
					lines              []change
				}{oldStart + 1, newStart + 1, nil})
				for k := start; k < idx; k++ {
					hunks[len(hunks)-1].lines = append(hunks[len(hunks)-1].lines, changes[k])
				}
				inHunk = true
			}
			currentLines = append(currentLines, c)
		} else if inHunk {
			currentLines = append(currentLines, c)
			changesAfter := 0
			for k := idx + 1; k < len(changes) && k <= idx+contextSize; k++ {
				if changes[k].kind != "equal" {
					changesAfter++
					break
				}
			}
			if changesAfter > 0 || len(currentLines) <= contextSize*2 {
				continue
			}
			hunks[len(hunks)-1].lines = append(hunks[len(hunks)-1].lines, currentLines[:contextSize]...)
			currentLines = nil
			inHunk = false
		}
	}
	if inHunk && len(currentLines) > 0 {
		hunks[len(hunks)-1].lines = append(hunks[len(hunks)-1].lines, currentLines...)
	}

	for _, h := range hunks {
		oldCount, newCount := 0, 0
		for _, c := range h.lines {
			switch c.kind {
			case "equal":
				oldCount++
				newCount++
			case "delete":
				oldCount++
			case "add":
				newCount++
			}
		}
		fmt.Printf("@@ -%d,%d +%d,%d @@\n", h.oldStart, oldCount, h.newStart, newCount)
		for _, c := range h.lines {
			switch c.kind {
			case "equal":
				fmt.Printf(" %s\n", c.line)
			case "delete":
				fmt.Printf("-%s\n", c.line)
			case "add":
				fmt.Printf("+%s\n", c.line)
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
