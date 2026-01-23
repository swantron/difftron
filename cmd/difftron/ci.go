package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swantron/difftron/internal/analyzer"
	"github.com/swantron/difftron/internal/coverage"
	"github.com/swantron/difftron/internal/hunk"
)

var (
	ciBaseRef    string
	ciHeadRef    string
	ciThreshold  float64
	ciOutputFile string
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run difftron in CI/CD environments",
	Long: `Run difftron analysis optimized for CI/CD pipelines.
This command handles git diff generation, coverage analysis, and
output formatting for GitHub Actions, GitLab CI, and other CI systems.`,
	RunE: runCI,
}

func init() {
	ciCmd.Flags().StringVar(&ciBaseRef, "base", "", "Base git ref (default: auto-detect from CI env)")
	ciCmd.Flags().StringVar(&ciHeadRef, "head", "", "Head git ref (default: auto-detect from CI env)")
	ciCmd.Flags().Float64Var(&ciThreshold, "threshold", 80.0, "Coverage threshold percentage")
	ciCmd.Flags().StringVar(&ciOutputFile, "output-file", "", "Output file for JSON results (default: stdout)")

	rootCmd.AddCommand(ciCmd)
}

func runCI(cmd *cobra.Command, args []string) error {
	// Get coverage file from args or env
	coverageFile := ""
	if len(args) > 0 {
		coverageFile = args[0]
	} else if envCoverage := os.Getenv("COVERAGE_FILE"); envCoverage != "" {
		coverageFile = envCoverage
	} else {
		coverageFile = "coverage.out"
	}

	// Auto-detect git refs from CI environment
	if ciBaseRef == "" {
		ciBaseRef = detectBaseRef()
	}
	if ciHeadRef == "" {
		ciHeadRef = detectHeadRef()
	}

	// Check if coverage file exists
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		return fmt.Errorf("coverage file not found: %s", coverageFile)
	}

	// Generate git diff
	diffOutput, err := getGitDiffForCI(ciBaseRef, ciHeadRef)
	if err != nil {
		// For direct pushes, if we can't get diff, that's OK - no changes to analyze
		fmt.Fprintf(os.Stderr, "Warning: Could not get git diff: %v\n", err)
		fmt.Println("No changes detected in diff.")
		return nil
	}

	if diffOutput == "" {
		fmt.Println("No changes detected in diff.")
		return nil
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

	// Parse coverage with better error handling
	format, err := coverage.DetectCoverageFormat(coverageFile)
	if err != nil {
		return fmt.Errorf("failed to detect coverage format: %w", err)
	}

	var coverageReport *coverage.Report
	if format == "go" {
		coverageReport, err = coverage.ParseGoCoverage(coverageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse as Go coverage (%v), trying LCOV format\n", err)
			// Fallback to LCOV
			coverageReport, err = coverage.ParseLCOV(coverageFile)
			if err != nil {
				return fmt.Errorf("failed to parse coverage file (tried both Go and LCOV): %w", err)
			}
		}
	} else {
		coverageReport, err = coverage.ParseLCOV(coverageFile)
		if err != nil {
			// Try Go format as fallback
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse as LCOV (%v), trying Go format\n", err)
			coverageReport, err = coverage.ParseGoCoverage(coverageFile)
			if err != nil {
				return fmt.Errorf("failed to parse coverage file (tried both LCOV and Go): %w", err)
			}
		}
	}

	// Analyze
	analysisResult, err := analyzer.Analyze(diffResult, coverageReport)
	if err != nil {
		return fmt.Errorf("failed to analyze: %w", err)
	}

	// Create CI output
	ciOutput := CIOutput{
		Coverage:        analysisResult.CoveragePercentage,
		Threshold:       ciThreshold,
		MeetsThreshold:  analysisResult.MeetsThreshold(ciThreshold),
		TotalLines:      analysisResult.TotalChangedLines,
		CoveredLines:    analysisResult.CoveredLines,
		UncoveredLines:  analysisResult.UncoveredLines,
		Files:           make(map[string]FileCIOutput),
	}

	for filePath, fileResult := range analysisResult.FileResults {
		ciOutput.Files[filePath] = FileCIOutput{
			Coverage:       fileResult.CoveragePercentage,
			CoveredLines:   fileResult.CoveredLines,
			UncoveredLines: fileResult.UncoveredLines,
			UncoveredLineNumbers: fileResult.UncoveredLineNumbers,
		}
	}

	// Output JSON
	jsonOutput, err := json.MarshalIndent(ciOutput, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if ciOutputFile != "" {
		if err := os.WriteFile(ciOutputFile, jsonOutput, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Results written to %s\n", ciOutputFile)
	} else {
		fmt.Println(string(jsonOutput))
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\n=== Difftron CI Analysis ===\n")
	fmt.Fprintf(os.Stderr, "Coverage: %.1f%% (threshold: %.1f%%)\n", 
		analysisResult.CoveragePercentage, ciThreshold)
	fmt.Fprintf(os.Stderr, "Status: %s\n", 
		map[bool]string{true: "PASS", false: "FAIL"}[ciOutput.MeetsThreshold])
	fmt.Fprintf(os.Stderr, "Changed Lines: %d | Covered: %d | Uncovered: %d\n",
		analysisResult.TotalChangedLines,
		analysisResult.CoveredLines,
		analysisResult.UncoveredLines)

	// Exit with appropriate code
	if !ciOutput.MeetsThreshold {
		os.Exit(1)
	}

	return nil
}

// CIOutput represents the structured output for CI systems
type CIOutput struct {
	Coverage       float64                  `json:"coverage_percentage"`
	Threshold      float64                  `json:"threshold"`
	MeetsThreshold bool                     `json:"meets_threshold"`
	TotalLines     int                      `json:"total_changed_lines"`
	CoveredLines   int                      `json:"covered_lines"`
	UncoveredLines int                      `json:"uncovered_lines"`
	Files          map[string]FileCIOutput `json:"files"`
}

// FileCIOutput represents file-level CI output
type FileCIOutput struct {
	Coverage            float64 `json:"coverage_percentage"`
	CoveredLines        int     `json:"covered_lines"`
	UncoveredLines      int     `json:"uncovered_lines"`
	UncoveredLineNumbers []int  `json:"uncovered_line_numbers"`
}

func detectBaseRef() string {
	// GitHub Actions
	if base := os.Getenv("GITHUB_BASE_REF"); base != "" {
		// For PRs, get the base SHA
		if sha := os.Getenv("GITHUB_BASE_SHA"); sha != "" {
			return sha
		}
		return base
	}
	if before := os.Getenv("GITHUB_EVENT_BEFORE"); before != "" {
		return before
	}

	// GitLab CI
	if base := os.Getenv("CI_MERGE_REQUEST_DIFF_BASE_SHA"); base != "" {
		return base
	}

	// Default: compare against previous commit
	return "HEAD~1"
}

func detectHeadRef() string {
	// GitHub Actions
	if head := os.Getenv("GITHUB_HEAD_SHA"); head != "" {
		return head
	}
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		return sha
	}

	// GitLab CI
	if head := os.Getenv("CI_COMMIT_SHA"); head != "" {
		return head
	}

	// Default: current HEAD
	return "HEAD"
}

func getGitDiffForCI(base, head string) (string, error) {
	// Handle special case where base might be a branch name
	if !strings.HasPrefix(base, "HEAD") && !isValidSHA(base) {
		// Try to get the SHA of the base branch
		cmd := exec.Command("git", "rev-parse", base)
		baseSHA, err := cmd.Output()
		if err == nil {
			base = strings.TrimSpace(string(baseSHA))
		}
	}

	cmd := exec.Command("git", "diff", base+".."+head)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(output), nil
}

func isValidSHA(s string) bool {
	// Basic SHA validation (40 chars for full SHA, 7+ for short)
	return len(s) >= 7 && len(s) <= 40 && isHex(s)
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
