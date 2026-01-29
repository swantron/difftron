package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/swantron/difftron/internal/analyzer"
	"github.com/swantron/difftron/internal/coverage"
	"github.com/swantron/difftron/internal/hunk"
	"github.com/swantron/difftron/pkg/report"
)

var (
	coverageFile      string
	diffFile          string
	threshold         float64
	thresholdNew      float64
	thresholdModified float64
	outputFormat      string
	baseRef           string
	headRef           string
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
	analyzeCmd.Flags().Float64VarP(&threshold, "threshold", "t", 80.0, "Coverage threshold percentage (applies to both new and modified files)")
	analyzeCmd.Flags().Float64Var(&thresholdNew, "threshold-new", 0, "Coverage threshold for new files (defaults to threshold if not set)")
	analyzeCmd.Flags().Float64Var(&thresholdModified, "threshold-modified", 0, "Coverage threshold for modified files (defaults to threshold if not set)")
	analyzeCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, markdown")
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
	switch format {
	case "go":
		// Parse Go's native coverage format
		coverageReport, err = coverage.ParseGoCoverage(coverageFile)
		if err != nil {
			// Fallback to LCOV if Go parsing fails
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse as Go coverage, trying LCOV format: %v\n", err)
			coverageReport, err = coverage.ParseLCOV(coverageFile)
			if err != nil {
				return fmt.Errorf("failed to parse coverage file (tried both Go and LCOV formats): %w", err)
			}
		}
	case "cobertura":
		// Parse Cobertura XML format
		coverageReport, err = coverage.ParseCobertura(coverageFile)
		if err != nil {
			return fmt.Errorf("failed to parse Cobertura coverage file: %w", err)
		}
	case "lcov":
		// Parse LCOV format
		coverageReport, err = coverage.ParseLCOV(coverageFile)
		if err != nil {
			return fmt.Errorf("failed to parse LCOV coverage file: %w", err)
		}
	default:
		return fmt.Errorf("unsupported coverage format: %s", format)
	}

	// Analyze
	analysisResult, err := analyzer.Analyze(diffResult, coverageReport)
	if err != nil {
		return fmt.Errorf("failed to analyze: %w", err)
	}

	// Set thresholds (use main threshold if specific ones not set)
	thresholdNew := thresholdNew
	thresholdModified := thresholdModified
	if thresholdNew == 0 {
		thresholdNew = threshold
	}
	if thresholdModified == 0 {
		thresholdModified = threshold
	}

	// Output results
	switch outputFormat {
	case "json":
		return outputJSON(analysisResult, thresholdNew, thresholdModified)
	case "markdown":
		return outputMarkdown(analysisResult, thresholdNew, thresholdModified)
	case "text":
		return outputText(analysisResult, thresholdNew, thresholdModified)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: text, json, markdown)", outputFormat)
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

func outputText(result *analyzer.AnalysisResult, thresholdNew, thresholdModified float64) error {
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

	// Show new vs modified breakdown if available
	if result.NewFileMetrics != nil && result.NewFileMetrics.FileCount > 0 {
		fmt.Printf("New Files Coverage: %.1f%% (%d files, %d/%d lines covered)\n",
			result.NewFileMetrics.CoveragePercentage,
			result.NewFileMetrics.FileCount,
			result.NewFileMetrics.CoveredLines,
			result.NewFileMetrics.TotalChangedLines)
	}
	if result.ModifiedFileMetrics != nil && result.ModifiedFileMetrics.FileCount > 0 {
		fmt.Printf("Modified Files Coverage: %.1f%% (%d files, %d/%d lines covered)\n",
			result.ModifiedFileMetrics.CoveragePercentage,
			result.ModifiedFileMetrics.FileCount,
			result.ModifiedFileMetrics.CoveredLines,
			result.ModifiedFileMetrics.TotalChangedLines)
	}
	fmt.Println()

	// Check thresholds
	meetsThresholds := result.MeetsThresholds(thresholdNew, thresholdModified)
	if meetsThresholds {
		fmt.Printf("✓ Coverage thresholds met\n")
		if thresholdNew != thresholdModified {
			fmt.Printf("  New files: %.1f%% >= %.1f%%\n", result.NewFileMetrics.CoveragePercentage, thresholdNew)
			fmt.Printf("  Modified files: %.1f%% >= %.1f%%\n", result.ModifiedFileMetrics.CoveragePercentage, thresholdModified)
		} else {
			fmt.Printf("  Overall: %.1f%% >= %.1f%%\n", result.CoveragePercentage, threshold)
		}
	} else {
		fmt.Printf("✗ Coverage thresholds not met\n")
		if result.NewFileMetrics != nil && result.NewFileMetrics.TotalChangedLines > 0 {
			if result.NewFileMetrics.CoveragePercentage < thresholdNew {
				fmt.Printf("  New files: %.1f%% < %.1f%%\n", result.NewFileMetrics.CoveragePercentage, thresholdNew)
			}
		}
		if result.ModifiedFileMetrics != nil && result.ModifiedFileMetrics.TotalChangedLines > 0 {
			if result.ModifiedFileMetrics.CoveragePercentage < thresholdModified {
				fmt.Printf("  Modified files: %.1f%% < %.1f%%\n", result.ModifiedFileMetrics.CoveragePercentage, thresholdModified)
			}
		}
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
	if !meetsThresholds {
		os.Exit(1)
	}

	return nil
}

func outputJSON(result *analyzer.AnalysisResult, thresholdNew, thresholdModified float64) error {
	// Use the higher threshold for JSON output (for backward compatibility)
	thresholdForJSON := threshold
	if thresholdNew > threshold {
		thresholdForJSON = thresholdNew
	}
	if thresholdModified > thresholdForJSON {
		thresholdForJSON = thresholdModified
	}

	jsonOutput, err := report.ToJSON(result, thresholdForJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Parse JSON and update meets_threshold to use the correct check
	var jsonData map[string]interface{}
	if err := json.Unmarshal(jsonOutput, &jsonData); err == nil {
		// Use MeetsThresholds for the actual check, not MeetsThreshold
		jsonData["meets_threshold"] = result.MeetsThresholds(thresholdNew, thresholdModified)
		jsonOutput, _ = json.MarshalIndent(jsonData, "", "  ")
	}

	fmt.Println(string(jsonOutput))

	if !result.MeetsThresholds(thresholdNew, thresholdModified) {
		os.Exit(1)
	}

	return nil
}

func outputMarkdown(result *analyzer.AnalysisResult, thresholdNew, thresholdModified float64) error {
	// Use the higher threshold for markdown output
	thresholdForMarkdown := threshold
	if thresholdNew > threshold {
		thresholdForMarkdown = thresholdNew
	}
	if thresholdModified > thresholdForMarkdown {
		thresholdForMarkdown = thresholdModified
	}

	markdownOutput := report.ToMarkdown(result, thresholdForMarkdown)
	fmt.Print(markdownOutput)

	if !result.MeetsThresholds(thresholdNew, thresholdModified) {
		os.Exit(1)
	}

	return nil
}
