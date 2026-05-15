package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loustack17/content-i18n/internal/config"
	"github.com/loustack17/content-i18n/internal/core"
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
		fmt.Printf("content-i18n mcp not yet implemented\nconfig: %s\n", *configPath)
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
		if arg == "status" || arg == "list" || arg == "plan" || arg == "apply-work" || arg == "validate-content" || arg == "validate-site" || arg == "mcp" || arg == "help" {
			rest := append([]string{}, args[:i]...)
			rest = append(rest, args[i+1:]...)
			return arg, rest
		}
	}

	return args[0], args[1:]
}

func printUsage() {
	fmt.Println("usage: content-i18n [--config path] <status|list|plan|apply-work|validate-content|validate-site|mcp>")
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

	if cfg.Adapter.Name != "hugo" {
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

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
