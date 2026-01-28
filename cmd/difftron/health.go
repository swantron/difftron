package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/swantron/difftron/internal/coverage"
	"github.com/swantron/difftron/internal/health"
	"github.com/swantron/difftron/internal/hunk"
)

var (
	healthUnitCoverage               string
	healthAPICoverage                string
	healthFunctionalCoverage         string
	healthBaselineUnitCoverage       string
	healthBaselineAPICoverage        string
	healthBaselineFunctionalCoverage string
	healthThreshold                  float64
	healthThresholdNew               float64
	healthThresholdModified          float64
	healthOutputFormat               string
	healthOutputFile                 string
	healthBaseRef                    string
	healthHeadRef                    string
	healthCommentPR                  bool
	healthCommentMR                  bool
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Comprehensive testing health analysis",
	Long: `Perform comprehensive testing health analysis by aggregating coverage
across multiple test types (unit, API, functional) and comparing against baseline.

This command provides holistic insights into testing health, including:
- Aggregated coverage across all test types
- Baseline comparison to detect regressions
- New vs modified file analysis
- Actionable insights and recommendations
- Support for PR/MR comments (GitHub/GitLab)`,
	RunE: runHealth,
}

func init() {
	healthCmd.Flags().StringVar(&healthUnitCoverage, "unit-coverage", "", "Path to unit test coverage file")
	healthCmd.Flags().StringVar(&healthAPICoverage, "api-coverage", "", "Path to API test coverage file")
	healthCmd.Flags().StringVar(&healthFunctionalCoverage, "functional-coverage", "", "Path to functional test coverage file")
	healthCmd.Flags().StringVar(&healthBaselineUnitCoverage, "baseline-unit-coverage", "", "Path to baseline unit test coverage file")
	healthCmd.Flags().StringVar(&healthBaselineAPICoverage, "baseline-api-coverage", "", "Path to baseline API test coverage file")
	healthCmd.Flags().StringVar(&healthBaselineFunctionalCoverage, "baseline-functional-coverage", "", "Path to baseline functional test coverage file")
	healthCmd.Flags().Float64Var(&healthThreshold, "threshold", 80.0, "Coverage threshold percentage")
	healthCmd.Flags().Float64Var(&healthThresholdNew, "threshold-new", 0, "Coverage threshold for new files (defaults to threshold if not set)")
	healthCmd.Flags().Float64Var(&healthThresholdModified, "threshold-modified", 0, "Coverage threshold for modified files (defaults to threshold if not set)")
	healthCmd.Flags().StringVarP(&healthOutputFormat, "output", "o", "text", "Output format: text, json, markdown")
	healthCmd.Flags().StringVar(&healthOutputFile, "output-file", "", "Output file path (default: stdout)")
	healthCmd.Flags().StringVar(&healthBaseRef, "base", "", "Base git ref for diff (default: auto-detect)")
	healthCmd.Flags().StringVar(&healthHeadRef, "head", "", "Head git ref for diff (default: auto-detect)")
	healthCmd.Flags().BoolVar(&healthCommentPR, "comment-pr", false, "Post comment on GitHub PR (requires GITHUB_TOKEN)")
	healthCmd.Flags().BoolVar(&healthCommentMR, "comment-mr", false, "Post comment on GitLab MR (requires GITLAB_TOKEN)")

	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	// Validate that at least one coverage file is provided
	if healthUnitCoverage == "" && healthAPICoverage == "" && healthFunctionalCoverage == "" {
		return fmt.Errorf("at least one coverage file is required (--unit-coverage, --api-coverage, or --functional-coverage)")
	}

	// Set thresholds (use main threshold if specific ones not set)
	if healthThresholdNew == 0 {
		healthThresholdNew = healthThreshold
	}
	if healthThresholdModified == 0 {
		healthThresholdModified = healthThreshold
	}

	// Get git diff
	baseRef := healthBaseRef
	headRef := healthHeadRef
	if baseRef == "" {
		baseRef = detectBaseRef()
	}
	if headRef == "" {
		headRef = detectHeadRef()
	}

	diffOutput, err := getGitDiffForPR(baseRef, headRef)
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
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

	// Load test coverage reports
	testReports := []*health.TestCoverageReport{}
	if healthUnitCoverage != "" {
		report, err := loadCoverageReport(healthUnitCoverage, health.TestTypeUnit)
		if err != nil {
			return fmt.Errorf("failed to load unit coverage: %w", err)
		}
		testReports = append(testReports, report)
	}
	if healthAPICoverage != "" {
		report, err := loadCoverageReport(healthAPICoverage, health.TestTypeAPI)
		if err != nil {
			return fmt.Errorf("failed to load API coverage: %w", err)
		}
		testReports = append(testReports, report)
	}
	if healthFunctionalCoverage != "" {
		report, err := loadCoverageReport(healthFunctionalCoverage, health.TestTypeFunctional)
		if err != nil {
			return fmt.Errorf("failed to load functional coverage: %w", err)
		}
		testReports = append(testReports, report)
	}

	// Load baseline reports if provided
	baselineReports := []*health.TestCoverageReport{}
	if healthBaselineUnitCoverage != "" {
		report, err := loadCoverageReport(healthBaselineUnitCoverage, health.TestTypeUnit)
		if err != nil {
			return fmt.Errorf("failed to load baseline unit coverage: %w", err)
		}
		baselineReports = append(baselineReports, report)
	}
	if healthBaselineAPICoverage != "" {
		report, err := loadCoverageReport(healthBaselineAPICoverage, health.TestTypeAPI)
		if err != nil {
			return fmt.Errorf("failed to load baseline API coverage: %w", err)
		}
		baselineReports = append(baselineReports, report)
	}
	if healthBaselineFunctionalCoverage != "" {
		report, err := loadCoverageReport(healthBaselineFunctionalCoverage, health.TestTypeFunctional)
		if err != nil {
			return fmt.Errorf("failed to load baseline functional coverage: %w", err)
		}
		baselineReports = append(baselineReports, report)
	}

	// Analyze health (use main threshold for now, will enhance with separate thresholds later)
	healthReport, err := health.AnalyzeHealth(diffResult, testReports, baselineReports, healthThreshold)
	if err != nil {
		return fmt.Errorf("failed to analyze health: %w", err)
	}

	// Generate output
	var output []byte
	switch healthOutputFormat {
	case "json":
		output, err = healthReport.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to generate JSON: %w", err)
		}
	case "markdown":
		output = []byte(healthReport.ToMarkdown())
	case "text":
		output = []byte(healthReport.ToStructuredText())
	default:
		return fmt.Errorf("unsupported output format: %s (supported: text, json, markdown)", healthOutputFormat)
	}

	// Write output
	if healthOutputFile != "" {
		if err := os.WriteFile(healthOutputFile, output, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Health report written to %s\n", healthOutputFile)
	} else {
		fmt.Print(string(output))
	}

	// Post PR/MR comments if requested
	if healthCommentPR {
		if err := postGitHubComment(healthReport); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to post GitHub comment: %v\n", err)
		}
	}
	if healthCommentMR {
		if err := postGitLabComment(healthReport); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to post GitLab comment: %v\n", err)
		}
	}

	// Determine exit code based on health status
	// Exit 1 if there are regressing files or overall coverage below threshold
	if healthReport.RegressingFiles > 0 || healthReport.ChangedCoverage < healthThreshold {
		os.Exit(1)
	}

	return nil
}

func loadCoverageReport(filePath string, testType health.TestType) (*health.TestCoverageReport, error) {
	format, err := coverage.DetectCoverageFormat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect coverage format: %w", err)
	}

	var coverageReport *coverage.Report
	switch format {
	case "go":
		coverageReport, err = coverage.ParseGoCoverage(filePath)
		if err != nil {
			// Try LCOV as fallback
			coverageReport, err = coverage.ParseLCOV(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse Go coverage (tried Go and LCOV): %w", err)
			}
		}
	case "cobertura":
		coverageReport, err = coverage.ParseCobertura(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Cobertura coverage: %w", err)
		}
	case "lcov":
		coverageReport, err = coverage.ParseLCOV(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LCOV coverage: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported coverage format: %s", format)
	}

	return &health.TestCoverageReport{
		TestType:       testType,
		CoverageReport: coverageReport,
		Source:         filePath,
	}, nil
}

func postGitHubComment(report *health.HealthReport) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Get PR number from environment
	prNumber := os.Getenv("GITHUB_PR_NUMBER")
	if prNumber == "" {
		// Try to extract from GITHUB_REF
		ref := os.Getenv("GITHUB_REF")
		if ref != "" {
			// GITHUB_REF format: refs/pull/:prNumber/merge
			// Extract PR number
		}
		if prNumber == "" {
			return fmt.Errorf("could not determine PR number (set GITHUB_PR_NUMBER)")
		}
	}

	repo := os.Getenv("GITHUB_REPOSITORY")
	if repo == "" {
		return fmt.Errorf("GITHUB_REPOSITORY environment variable not set")
	}

	// Generate markdown comment
	comment := report.ToMarkdown()

	// TODO: Implement GitHub API call to post comment
	// For now, just print a message
	fmt.Fprintf(os.Stderr, "GitHub PR comment would be posted to %s PR #%s\n", repo, prNumber)
	fmt.Fprintf(os.Stderr, "Comment preview:\n%s\n", comment)

	return nil
}

func postGitLabComment(report *health.HealthReport) error {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITLAB_TOKEN environment variable not set")
	}

	mrIID := os.Getenv("CI_MERGE_REQUEST_IID")
	if mrIID == "" {
		return fmt.Errorf("CI_MERGE_REQUEST_IID environment variable not set")
	}

	projectID := os.Getenv("CI_PROJECT_ID")
	if projectID == "" {
		return fmt.Errorf("CI_PROJECT_ID environment variable not set")
	}

	// Generate markdown comment
	comment := report.ToMarkdown()

	// TODO: Implement GitLab API call to post comment
	// For now, just print a message
	fmt.Fprintf(os.Stderr, "GitLab MR comment would be posted to project %s MR !%s\n", projectID, mrIID)
	fmt.Fprintf(os.Stderr, "Comment preview:\n%s\n", comment)

	return nil
}
