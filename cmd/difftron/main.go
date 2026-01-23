package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	rootCmd = &cobra.Command{
		Use:   "difftron",
		Short: "AI-powered Quality Gate CLI for code coverage analysis",
		Long: `Difftron is a language-agnostic, AI-powered Quality Gate CLI.
It ensures that new code changes are adequately tested by correlating
git diff hunks with standard coverage reports (LCOV, Cobertura, etc.).`,
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
	}
)

func main() {
	// Subcommands are added in their respective files via init() functions

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
