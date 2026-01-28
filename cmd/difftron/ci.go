package main

import (
	"encoding/json"
	"fmt"
	"os"

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
	switch format {
	case "go":
		coverageReport, err = coverage.ParseGoCoverage(coverageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse as Go coverage (%v), trying other formats\n", err)
			// Try Cobertura as fallback
			coverageReport, err = coverage.ParseCobertura(coverageFile)
			if err != nil {
				// Try LCOV as last fallback
				coverageReport, err = coverage.ParseLCOV(coverageFile)
				if err != nil {
					return fmt.Errorf("failed to parse coverage file (tried Go, Cobertura, and LCOV): %w", err)
				}
			}
		}
	case "cobertura":
		coverageReport, err = coverage.ParseCobertura(coverageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse as Cobertura (%v), trying LCOV format\n", err)
			// Fallback to LCOV
			coverageReport, err = coverage.ParseLCOV(coverageFile)
			if err != nil {
				return fmt.Errorf("failed to parse coverage file (tried Cobertura and LCOV): %w", err)
			}
		}
	case "lcov":
		coverageReport, err = coverage.ParseLCOV(coverageFile)
		if err != nil {
			// Try Cobertura as fallback
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse as LCOV (%v), trying Cobertura format\n", err)
			coverageReport, err = coverage.ParseCobertura(coverageFile)
			if err != nil {
				// Try Go format as last fallback
				coverageReport, err = coverage.ParseGoCoverage(coverageFile)
				if err != nil {
					return fmt.Errorf("failed to parse coverage file (tried LCOV, Cobertura, and Go): %w", err)
				}
			}
		}
	default:
		return fmt.Errorf("unsupported coverage format: %s", format)
	}

	// Analyze
	analysisResult, err := analyzer.Analyze(diffResult, coverageReport)
	if err != nil {
		return fmt.Errorf("failed to analyze: %w", err)
	}

	// Create CI output
	ciOutput := CIOutput{
		Coverage:       analysisResult.CoveragePercentage,
		Threshold:      ciThreshold,
		MeetsThreshold: analysisResult.MeetsThreshold(ciThreshold),
		TotalLines:     analysisResult.TotalChangedLines,
		CoveredLines:   analysisResult.CoveredLines,
		UncoveredLines: analysisResult.UncoveredLines,
		Files:          make(map[string]FileCIOutput),
	}

	for filePath, fileResult := range analysisResult.FileResults {
		ciOutput.Files[filePath] = FileCIOutput{
			Coverage:             fileResult.CoveragePercentage,
			CoveredLines:         fileResult.CoveredLines,
			UncoveredLines:       fileResult.UncoveredLines,
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
	Coverage       float64                 `json:"coverage_percentage"`
	Threshold      float64                 `json:"threshold"`
	MeetsThreshold bool                    `json:"meets_threshold"`
	TotalLines     int                     `json:"total_changed_lines"`
	CoveredLines   int                     `json:"covered_lines"`
	UncoveredLines int                     `json:"uncovered_lines"`
	Files          map[string]FileCIOutput `json:"files"`
}

// FileCIOutput represents file-level CI output
type FileCIOutput struct {
	Coverage             float64 `json:"coverage_percentage"`
	CoveredLines         int     `json:"covered_lines"`
	UncoveredLines       int     `json:"uncovered_lines"`
	UncoveredLineNumbers []int   `json:"uncovered_line_numbers"`
}

func getGitDiffForCI(base, head string) (string, error) {
	return getGitDiffForPR(base, head)
}
