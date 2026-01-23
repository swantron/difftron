package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/swantron/difftron/internal/analyzer"
	"github.com/swantron/difftron/internal/coverage"
	"github.com/swantron/difftron/internal/hunk"
)

var (
	coverageFile string
	diffFile     string
	threshold    float64
	outputFormat string
	baseRef      string
	headRef      string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze git diff against coverage data",
	Long: `Analyze git diff hunks against coverage reports to identify
uncovered lines in your changes.`,
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().StringVarP(&coverageFile, "coverage", "c", "", "Path to coverage file (LCOV format)")
	analyzeCmd.Flags().StringVarP(&diffFile, "diff", "d", "", "Path to git diff file (optional, uses git diff if not provided)")
	analyzeCmd.Flags().Float64VarP(&threshold, "threshold", "t", 80.0, "Coverage threshold percentage")
	analyzeCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json")
	analyzeCmd.Flags().StringVarP(&baseRef, "base", "b", "HEAD", "Base ref for git diff (default: HEAD)")
	analyzeCmd.Flags().StringVarP(&headRef, "head", "", "HEAD", "Head ref for git diff (default: HEAD)")

	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Validate coverage file
	if coverageFile == "" {
		return fmt.Errorf("coverage file is required (use --coverage or -c)")
	}

	// Get git diff
	var diffOutput string
	var err error

	if diffFile != "" {
		// Read from file
		file, err := os.Open(diffFile)
		if err != nil {
			return fmt.Errorf("failed to open diff file: %w", err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read diff file: %w", err)
		}
		diffOutput = string(content)
	} else {
		// Get diff from git
		diffOutput, err = getGitDiff(baseRef, headRef)
		if err != nil {
			return fmt.Errorf("failed to get git diff: %w", err)
		}
	}

	// Parse diff
	diffResult, err := hunk.ParseGitDiff(diffOutput)
	if err != nil {
		return fmt.Errorf("failed to parse git diff: %w", err)
	}

	if !diffResult.HasChanges() {
		fmt.Println("No changes detected in diff.")
		return nil
	}

	// Detect coverage format and parse
	format, err := coverage.DetectCoverageFormat(coverageFile)
	if err != nil {
		return fmt.Errorf("failed to detect coverage format: %w", err)
	}

	var coverageReport *coverage.Report
	if format == "go" {
		// Parse Go's native coverage format
		coverageReport, err = coverage.ParseGoCoverage(coverageFile)
		if err != nil {
			return fmt.Errorf("failed to parse Go coverage file: %w", err)
		}
	} else {
		// Parse LCOV format
		coverageReport, err = coverage.ParseLCOV(coverageFile)
		if err != nil {
			return fmt.Errorf("failed to parse LCOV coverage file: %w", err)
		}
	}

	// Analyze
	analysisResult, err := analyzer.Analyze(diffResult, coverageReport)
	if err != nil {
		return fmt.Errorf("failed to analyze: %w", err)
	}

	// Output results
	switch outputFormat {
	case "json":
		return outputJSON(analysisResult)
	case "text":
		return outputText(analysisResult)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func getGitDiff(base, head string) (string, error) {
	// If base and head are the same, get diff of working directory
	if base == head && base == "HEAD" {
		cmd := exec.Command("git", "diff", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			// Try staged changes
			cmd = exec.Command("git", "diff", "--cached")
			output, err = cmd.Output()
			if err != nil {
				return "", fmt.Errorf("failed to get git diff: %w", err)
			}
		}
		return string(output), nil
	}

	// Get diff between two refs
	cmd := exec.Command("git", "diff", base, head)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}
	return string(output), nil
}

func outputText(result *analyzer.AnalysisResult) error {
	fmt.Println("Difftron Coverage Analysis")
	fmt.Println("==========================")
	fmt.Println()

	if result.TotalChangedLines == 0 {
		fmt.Println("No changed lines to analyze.")
		return nil
	}

	fmt.Printf("Overall Coverage: %.1f%% (%d/%d lines covered)\n",
		result.CoveragePercentage,
		result.CoveredLines,
		result.TotalChangedLines)
	fmt.Println()

	// Check threshold
	meetsThreshold := result.MeetsThreshold(threshold)
	if meetsThreshold {
		fmt.Printf("✓ Coverage threshold met (%.1f%% >= %.1f%%)\n", result.CoveragePercentage, threshold)
	} else {
		fmt.Printf("✗ Coverage threshold not met (%.1f%% < %.1f%%)\n", result.CoveragePercentage, threshold)
	}
	fmt.Println()

	// Per-file results
	fmt.Println("Per-File Results:")
	fmt.Println("-----------------")
	for filePath, fileResult := range result.FileResults {
		fmt.Printf("\n%s\n", filePath)
		fmt.Printf("  Coverage: %.1f%% (%d/%d lines covered)\n",
			fileResult.CoveragePercentage,
			fileResult.CoveredLines,
			fileResult.TotalChangedLines)

		if len(fileResult.UncoveredLineNumbers) > 0 {
			fmt.Printf("  Uncovered lines: %v\n", fileResult.UncoveredLineNumbers)
		}
	}

	// Exit with error if threshold not met
	if !meetsThreshold {
		os.Exit(1)
	}

	return nil
}

func outputJSON(result *analyzer.AnalysisResult) error {
	// Simple JSON output (can be enhanced with proper JSON marshaling)
	fmt.Printf(`{
  "total_changed_lines": %d,
  "covered_lines": %d,
  "uncovered_lines": %d,
  "coverage_percentage": %.2f,
  "meets_threshold": %t,
  "files": {
`,
		result.TotalChangedLines,
		result.CoveredLines,
		result.UncoveredLines,
		result.CoveragePercentage,
		result.MeetsThreshold(threshold))

	first := true
	for filePath, fileResult := range result.FileResults {
		if !first {
			fmt.Print(",\n")
		}
		first = false
		fmt.Printf(`    "%s": {
      "coverage_percentage": %.2f,
      "covered_lines": %d,
      "uncovered_lines": %d,
      "uncovered_line_numbers": %v
    }`,
			filePath,
			fileResult.CoveragePercentage,
			fileResult.CoveredLines,
			fileResult.UncoveredLines,
			fileResult.UncoveredLineNumbers)
	}

	fmt.Println("\n  }\n}")

	if !result.MeetsThreshold(threshold) {
		os.Exit(1)
	}

	return nil
}
