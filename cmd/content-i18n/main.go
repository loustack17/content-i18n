package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loustack/content-i18n/internal/config"
	"github.com/loustack/content-i18n/internal/content"
)

func main() {
	command, args := parseCommand(os.Args[1:])
	flags := flag.NewFlagSet(command, flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	switch command {
	case "status":
		runStatus(*configPath)
	case "list":
		runList(*configPath)
	case "plan":
		runPlan(*configPath, flags)
	case "translate":
		runTranslate(*configPath, flags)
	case "validate":
		runValidate(*configPath, flags)
	case "validate-site":
		runValidateSite(*configPath)
	case "mcp":
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
		if arg == "status" || arg == "list" || arg == "plan" || arg == "translate" || arg == "validate" || arg == "validate-site" || arg == "mcp" || arg == "help" {
			rest := append([]string{}, args[:i]...)
			rest = append(rest, args[i+1:]...)
			return arg, rest
		}
	}

	return args[0], args[1:]
}

func printUsage() {
	fmt.Println("usage: content-i18n [--config path] <status|list|plan|translate|validate|validate-site|mcp>")
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

	files, err := content.Discover(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	report := content.BuildStatusReport(cfg, files)
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

	files, err := content.Discover(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-60s %-10s %-60s %s\n", "source_path", "language", "target_path", "status")
	for _, f := range files {
		fmt.Printf("%-60s %-10s %-60s %s\n", f.SourcePath, f.Language, f.TargetPath, f.Status)
	}
}

func runPlan(configPath string, flags *flag.FlagSet) {
	file := flags.String("file", "", "source file to plan")
	to := flags.String("to", "", "target language")
	flags.Parse(flags.Args())

	if *file == "" || *to == "" {
		fmt.Fprintln(os.Stderr, "usage: content-i18n plan --file <source.md> --to <lang>")
		os.Exit(2)
	}

	fmt.Printf("plan: %s -> %s\n", *file, *to)
	fmt.Println("not yet implemented")
}

func runTranslate(configPath string, flags *flag.FlagSet) {
	file := flags.String("file", "", "source file to translate")
	to := flags.String("to", "", "target language")
	provider := flags.String("provider", "ai-harness", "translation provider")
	flags.Parse(flags.Args())

	if *file == "" || *to == "" {
		fmt.Fprintln(os.Stderr, "usage: content-i18n translate --file <source.md> --to <lang> [--provider <provider>]")
		os.Exit(2)
	}

	fmt.Printf("translate: %s -> %s (provider: %s)\n", *file, *to, *provider)
	fmt.Println("not yet implemented")
}

func runValidate(configPath string, flags *flag.FlagSet) {
	file := flags.String("file", "", "target file to validate")
	flags.Parse(flags.Args())

	if *file == "" {
		fmt.Fprintln(os.Stderr, "usage: content-i18n validate --file <target.md>")
		os.Exit(2)
	}

	fmt.Printf("validate: %s\n", *file)
	fmt.Println("not yet implemented")
}

func runValidateSite(configPath string) {
	fmt.Println("validate-site: not yet implemented")
}
