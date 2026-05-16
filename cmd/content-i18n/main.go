package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/core"
	"github.com/loustack17/content-i18n/internal/mcp"
)

func main() {
	command, args := parseCommand(os.Args[1:])

	switch command {
	case "status":
		flags := flag.NewFlagSet(command, flag.ExitOnError)
		configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
		flags.Parse(args)
		runStatus(*configPath)
	case "list":
		flags := flag.NewFlagSet(command, flag.ExitOnError)
		configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
		flags.Parse(args)
		runList(*configPath)
	case "plan":
		runPlan(args)
	case "apply-work":
		runApplyWork(args)
	case "validate-content":
		runValidateContent(args)
	case "validate-site":
		flags := flag.NewFlagSet(command, flag.ExitOnError)
		configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
		flags.Parse(args)
		runValidateSite(*configPath)
	case "mcp":
		flags := flag.NewFlagSet(command, flag.ExitOnError)
		configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
		flags.Parse(args)
		runMCP(*configPath)
	case "prepare":
		runPrepare(args)
	case "review":
		runReview(args)
	case "repair-plan":
		runRepairPlan(args)
	case "next":
		runNext(args)
	case "batch-status":
		runBatchStatus(args)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(2)
	}
}

func parseCommand(args []string) (string, []string) {
	if len(args) == 0 {
		return "help", nil
	}

	for i, arg := range args {
		if arg == "status" || arg == "list" || arg == "plan" || arg == "apply-work" || arg == "validate-content" || arg == "validate-site" || arg == "mcp" || arg == "prepare" || arg == "review" || arg == "repair-plan" || arg == "next" || arg == "batch-status" || arg == "help" {
			rest := append([]string{}, args[:i]...)
			rest = append(rest, args[i+1:]...)
			return arg, rest
		}
	}

	return args[0], args[1:]
}

func printUsage() {
	fmt.Println("usage: content-i18n [--config path] <status|list|plan|apply-work|validate-content|validate-site|prepare|review|repair-plan|next|batch-status|mcp>")
}

func loadConfig(configPath string) (*config.Config, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}
	return config.Load(absPath)
}

func runStatus(configPath string) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	report, err := core.Status(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("project_type: %s\n", report.ProjectType)
	fmt.Printf("source_language: %s\n", report.SourceLanguage)
	fmt.Printf("source_path: %s\n", report.SourcePath)
	fmt.Printf("target_languages: %s\n", strings.Join(report.TargetLanguages, ", "))
	for lang, path := range report.TargetPaths {
		fmt.Printf("target_path[%s]: %s\n", lang, path)
	}
	fmt.Printf("source_file_count: %d\n", report.SourceFileCount)
	fmt.Printf("target_file_count: %d\n", report.TargetFileCount)
	fmt.Printf("missing_translation_count: %d\n", report.MissingCount)
}

func runList(configPath string) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	files, err := core.List(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-60s %-10s %-60s %s\n", "source_path", "language", "target_path", "status")
	for _, f := range files {
		fmt.Printf("%-60s %-10s %-60s %s\n", f.SourcePath, f.Language, f.TargetPath, f.Status)
	}
}

func runPlan(args []string) {
	flags := flag.NewFlagSet("plan", flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	file := flags.String("file", "", "source file to plan")
	to := flags.String("to", "", "target language")
	flags.Parse(args)

	cfg, err := loadConfig(*configPath)
	exitOnError(err)

	var plans []core.FileInfo
	if (*file != "") != (*to != "") {
		fmt.Fprintln(os.Stderr, "error: --file and --to must be used together")
		os.Exit(2)
	}
	if *file != "" && *to != "" {
		absFile, err := filepath.Abs(*file)
		exitOnError(err)
		plans, err = core.Plan(cfg, absFile, *to)
		exitOnError(err)
	} else {
		plans, err = core.Plan(cfg, "", "")
		exitOnError(err)
	}

	for _, p := range plans {
		fmt.Printf("%s -> %s [%s] %s\n", p.SourcePath, p.Language, p.Status, p.TargetPath)
	}
}

func runApplyWork(args []string) {
	flags := flag.NewFlagSet("apply-work", flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	slug := flags.String("slug", "", "content slug to translate")
	dryRun := flags.Bool("dry-run", false, "show plan without executing")
	force := flags.Bool("force", false, "override validation errors")
	flags.Parse(args)

	if *slug == "" {
		fmt.Fprintln(os.Stderr, "usage: content-i18n apply-work --slug <slug> [--dry-run] [--force]")
		os.Exit(2)
	}

	cfg, err := loadConfig(*configPath)
	exitOnError(err)

	err = core.ApplyWork(cfg, *slug, *dryRun, *force)
	exitOnError(err)
}

func runValidateContent(args []string) {
	flags := flag.NewFlagSet("validate-content", flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	file := flags.String("file", "", "target file to validate")
	source := flags.String("source", "", "source file for comparison")
	glossary := flags.String("glossary", "", "glossary file path")
	flags.Parse(args)

	if *file == "" {
		fmt.Fprintln(os.Stderr, "usage: content-i18n validate-content --file <target.md> [--source <source.md>] [--glossary <path>]")
		os.Exit(2)
	}

	cfg, err := loadConfig(*configPath)
	exitOnError(err)

	opts := &core.ValidateOptions{
		SourcePath:   *source,
		GlossaryPath: *glossary,
		Config:       cfg,
	}

	result, err := core.ValidateContent(*file, opts)
	exitOnError(err)

	fmt.Printf("validate-content: %s\n", *file)
	if result.Passed {
		fmt.Println("PASS")
	} else {
		fmt.Println("FAIL")
		for _, v := range result.Violations {
			fmt.Printf("  [%s] %s: %s (fix: %s)\n", v.Field, v.Section, v.Message, v.SuggestedFix)
		}
		os.Exit(1)
	}
}

func runValidateSite(configPath string) {
	cfg, err := loadConfig(configPath)
	exitOnError(err)

	if cfg.Adapter.Name != core.AdapterHugo {
		fmt.Fprintf(os.Stderr, "error: validate-site only supports hugo adapter (got: %s)\n", cfg.Adapter.Name)
		os.Exit(2)
	}

	hugoRoot := cfg.ConfigDir

	warnings := core.ValidateSiteConfig(cfg)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	result, err := core.ValidateSite(cfg, hugoRoot)
	exitOnError(err)

	fmt.Printf("validate-site: hugo output: %s\n", result.HugoOutput)
	if result.Passed {
		fmt.Println("PASS")
	} else {
		fmt.Println("FAIL")
		for _, v := range result.Violations {
			fmt.Printf("  %s\n", v)
		}
		os.Exit(1)
	}
}

func runMCP(configPath string) {
	cfg, err := loadConfig(configPath)
	exitOnError(err)

	srv := mcp.NewServer(cfg, configPath)
	if err := srv.ServeStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runPrepare(args []string) {
	flags := flag.NewFlagSet("prepare", flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	file := flags.String("file", "", "source file to prepare for translation")
	to := flags.String("to", "", "target language")
	flags.Parse(args)

	if *file == "" || *to == "" {
		fmt.Fprintln(os.Stderr, "usage: content-i18n prepare --file <source.md> --to <lang>")
		os.Exit(2)
	}

	cfg, err := loadConfig(*configPath)
	exitOnError(err)

	absFile, err := filepath.Abs(*file)
	exitOnError(err)

	result, err := core.TranslatePrepare(cfg, absFile, *to)
	exitOnError(err)

	fmt.Printf("slug: %s\n", result.Slug)
	fmt.Printf("target_path: %s\n", result.TargetPath)
	fmt.Printf("\n--- source ---\n%s\n", result.Source)
	if result.Prompt != "" {
		fmt.Printf("\n--- prompt ---\n%s\n", result.Prompt)
	}
	if result.Glossary != "" {
		fmt.Printf("\n--- glossary ---\n%s\n", result.Glossary)
	}
	if result.Style != "" {
		fmt.Printf("\n--- style ---\n%s\n", result.Style)
	}
	if result.Context != "" {
		fmt.Printf("\n--- context ---\n%s\n", result.Context)
	}
}

func runReview(args []string) {
	flags := flag.NewFlagSet("review", flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	file := flags.String("file", "", "target file to review")
	source := flags.String("source", "", "source file for comparison")
	flags.Parse(args)

	if *file == "" || *source == "" {
		fmt.Fprintln(os.Stderr, "usage: content-i18n review --file <target.md> --source <source.md>")
		os.Exit(2)
	}

	cfg, err := loadConfig(*configPath)
	exitOnError(err)

	result, err := core.TranslateReview(cfg, *source, *file)
	exitOnError(err)

	fmt.Printf("source_words: %d\n", result.SourceWords)
	fmt.Printf("target_words: %d\n", result.TargetWords)
	fmt.Printf("word_ratio: %s\n", result.WordRatio)
	fmt.Printf("passed: %v\n", result.Passed)

	if len(result.Issues) > 0 {
		fmt.Println("\nissues:")
		for _, issue := range result.Issues {
			fmt.Printf("  [%s] [%s] %s: %s (fix: %s)\n", issue.Severity, issue.Field, issue.Section, issue.Message, issue.SuggestedFix)
		}
	}

	if !result.Passed {
		os.Exit(1)
	}
}

func runRepairPlan(args []string) {
	flags := flag.NewFlagSet("repair-plan", flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	file := flags.String("file", "", "target file with repaired content")
	source := flags.String("source", "", "source file for comparison")
	flags.Parse(args)

	if *file == "" || *source == "" {
		fmt.Fprintln(os.Stderr, "usage: content-i18n repair-plan --file <target.md> --source <source.md>")
		os.Exit(2)
	}

	cfg, err := loadConfig(*configPath)
	exitOnError(err)

	result, err := core.TranslateReview(cfg, *source, *file)
	exitOnError(err)

	if result.Passed {
		fmt.Println("REPAIR OK")
		fmt.Printf("source_words: %d, target_words: %d, ratio: %s\n", result.SourceWords, result.TargetWords, result.WordRatio)
	} else {
		fmt.Println("REPAIR FAILED")
		for _, issue := range result.Issues {
			fmt.Printf("  [%s] [%s] %s: %s\n", issue.Severity, issue.Field, issue.Section, issue.Message)
		}
		os.Exit(1)
	}
}

func runNext(args []string) {
	flags := flag.NewFlagSet("next", flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	group := flags.String("group", "", "filter by group name (e.g. DevOps)")
	flags.Parse(args)

	cfg, err := loadConfig(*configPath)
	exitOnError(err)

	entry, err := core.NextTranslation(cfg, *group)
	exitOnError(err)

	if entry == nil {
		fmt.Println("queue empty")
		return
	}

	fmt.Printf("source: %s\n", entry.SourcePath)
	fmt.Printf("target: %s\n", entry.TargetPath)
	fmt.Printf("language: %s\n", entry.Language)
	fmt.Printf("status: %s\n", entry.Status)
	fmt.Printf("source_hash: %s\n", entry.SourceHash)
}

func runBatchStatus(args []string) {
	flags := flag.NewFlagSet("batch-status", flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	group := flags.String("group", "", "filter by group name (e.g. DevOps)")
	flags.Parse(args)

	cfg, err := loadConfig(*configPath)
	exitOnError(err)

	status, err := core.TranslationQueue(cfg, *group)
	exitOnError(err)

	fmt.Printf("total: %d\n", status.Total)
	fmt.Printf("completed: %d\n", status.Completed)
	fmt.Printf("stale: %d\n", status.Stale)
	fmt.Printf("missing: %d\n", status.Missing)
	if status.Next != nil {
		fmt.Printf("next: %s [%s]\n", status.Next.SourcePath, status.Next.Language)
	} else {
		fmt.Println("next: none")
	}
}
