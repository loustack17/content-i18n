package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	command, args := parseCommand(os.Args[1:])
	flags := flag.NewFlagSet(command, flag.ExitOnError)
	configPath := flags.String("config", "content-i18n.yaml", "path to content-i18n config")
	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	switch command {
	case "status", "list", "plan", "translate", "validate", "validate-site", "mcp":
		fmt.Printf("content-i18n %s\nconfig: %s\n", command, *configPath)
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
