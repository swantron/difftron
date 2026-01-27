package health

import (
	"fmt"

	"github.com/swantron/difftron/internal/coverage"
	"github.com/swantron/difftron/internal/hunk"
)

// TestType represents different types of tests
type TestType string

const (
	TestTypeUnit        TestType = "unit"
	TestTypeAPI         TestType = "api"
	TestTypeFunctional  TestType = "functional"
	TestTypeIntegration TestType = "integration"
	TestTypeE2E         TestType = "e2e"
)

// TestCoverageReport represents coverage from a specific test type
type TestCoverageReport struct {
	TestType       TestType
	CoverageReport *coverage.Report
	TestCount      int    // Number of tests run
	Duration       string // Test execution duration
	Source         string // Source file or command that generated this coverage
}

// FileHealth represents comprehensive health metrics for a single file
type FileHealth struct {
	FilePath string

	// Overall metrics (aggregated across all test types)
	TotalLines         int
	CoveredLines       int
	UncoveredLines     int
	CoveragePercentage float64

	// Per-test-type coverage
	UnitTestCoverage       float64
	APITestCoverage        float64
	FunctionalTestCoverage float64

	// Change-specific metrics
	ChangedLines              int
	ChangedCoveredLines       int
	ChangedUncoveredLines     int
	ChangedCoveragePercentage float64

	// Baseline comparison
	BaselineCoveragePercentage float64
	CoverageDelta              float64 // Current - Baseline

	// Test type breakdown for changed lines
	ChangedLinesCoveredByUnit       int
	ChangedLinesCoveredByAPI        int
	ChangedLinesCoveredByFunctional int

	// Health indicators
	IsNewFile      bool
	HasRegression  bool // Coverage dropped below baseline
	NeedsAttention bool // Low coverage or regression
}

// HealthReport provides a comprehensive view of testing health
type HealthReport struct {
	// Overall project metrics
	TotalFiles          int
	TotalLines          int
	TotalCoveredLines   int
	TotalUncoveredLines int
	OverallCoverage     float64

	// Change-specific metrics
	ChangedFiles          int
	ChangedLines          int
	ChangedCoveredLines   int
	ChangedUncoveredLines int
	ChangedCoverage       float64

	// Test type coverage summary
	UnitTestCoverage       float64
	APITestCoverage        float64
	FunctionalTestCoverage float64

	// File-level health
	FileHealth map[string]*FileHealth

	// New vs Modified breakdown
	NewFilesCount         int
	ModifiedFilesCount    int
	NewFilesCoverage      float64
	ModifiedFilesCoverage float64

	// Health summary
	HealthyFiles    int // Files meeting threshold
	AtRiskFiles     int // Files below threshold
	RegressingFiles int // Files with coverage drop

	// Insights and recommendations
	Insights        []Insight
	Recommendations []Recommendation
}

// Insight provides actionable information about testing health
type Insight struct {
	Type        string // "info", "warning", "error", "success"
	Category    string // "coverage", "regression", "test-gap", "new-file"
	Title       string
	Description string
	File        string // Optional: specific file this relates to
	LineNumbers []int  // Optional: specific lines
	Severity    string // "low", "medium", "high", "critical"
}

// Recommendation suggests actions to improve testing health
type Recommendation struct {
	Priority    string // "low", "medium", "high", "critical"
	Category    string // "add-tests", "improve-coverage", "fix-regression"
	Title       string
	Description string
	Action      string   // Specific action to take
	Files       []string // Files that need attention
	TestType    TestType // Recommended test type
}

// AggregateCoverage combines multiple test type coverage reports
// Lines are considered covered if ANY test type covers them
func AggregateCoverage(reports []*TestCoverageReport) (*coverage.Report, error) {
	if len(reports) == 0 {
		return &coverage.Report{
			FileCoverage: make(map[string]*coverage.CoverageData),
		}, nil
	}

	aggregated := &coverage.Report{
		FileCoverage: make(map[string]*coverage.CoverageData),
	}

	// Aggregate coverage across all test types
	// A line is covered if ANY test type covers it
	for _, testReport := range reports {
		if testReport.CoverageReport == nil {
			continue
		}

		for filePath, fileCoverage := range testReport.CoverageReport.FileCoverage {
			// Initialize aggregated coverage for this file if needed
			if aggregated.FileCoverage[filePath] == nil {
				aggregated.FileCoverage[filePath] = &coverage.CoverageData{
					LineHits: make(map[int]int),
				}
			}

			aggFileCoverage := aggregated.FileCoverage[filePath]

			// Merge line hits - take maximum hit count across test types
			for lineNum, hits := range fileCoverage.LineHits {
				currentHits := aggFileCoverage.LineHits[lineNum]
				if hits > currentHits {
					aggFileCoverage.LineHits[lineNum] = hits
				}
			}

			// Update totals
			aggFileCoverage.TotalLines = fileCoverage.TotalLines
			aggFileCoverage.CoveredLines = 0
			for _, hits := range aggFileCoverage.LineHits {
				if hits > 0 {
					aggFileCoverage.CoveredLines++
				}
			}
		}
	}

	return aggregated, nil
}

// AnalyzeHealth provides comprehensive testing health analysis
func AnalyzeHealth(
	diffResult *hunk.ParseResult,
	testReports []*TestCoverageReport,
	baselineReports []*TestCoverageReport,
	threshold float64,
) (*HealthReport, error) {
	if diffResult == nil {
		return nil, fmt.Errorf("diff result cannot be nil")
	}
	if len(testReports) == 0 {
		return nil, fmt.Errorf("at least one test coverage report is required")
	}

	// Aggregate current coverage across all test types
	currentAggregated, err := AggregateCoverage(testReports)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate coverage: %w", err)
	}

	// Aggregate baseline coverage if provided
	var baselineAggregated *coverage.Report
	if len(baselineReports) > 0 {
		baselineAggregated, err = AggregateCoverage(baselineReports)
		if err != nil {
			return nil, fmt.Errorf("failed to aggregate baseline coverage: %w", err)
		}
	}

	report := &HealthReport{
		FileHealth:      make(map[string]*FileHealth),
		Insights:        make([]Insight, 0),
		Recommendations: make([]Recommendation, 0),
	}

	// Calculate overall project metrics
	report.calculateOverallMetrics(currentAggregated)

	// Analyze each changed file
	for filePath, changedLines := range diffResult.ChangedLines {
		fileHealth := report.analyzeFileHealth(
			filePath,
			changedLines,
			diffResult,
			testReports,
			baselineAggregated,
			threshold,
		)
		report.FileHealth[filePath] = fileHealth

		// Update change-specific metrics
		report.ChangedFiles++
		report.ChangedLines += fileHealth.ChangedLines
		report.ChangedCoveredLines += fileHealth.ChangedCoveredLines
		report.ChangedUncoveredLines += fileHealth.ChangedUncoveredLines

		if fileHealth.IsNewFile {
			report.NewFilesCount++
		} else {
			report.ModifiedFilesCount++
		}
	}

	// Calculate change coverage
	if report.ChangedLines > 0 {
		report.ChangedCoverage = float64(report.ChangedCoveredLines) / float64(report.ChangedLines) * 100
	}

	// Calculate new vs modified coverage
	report.calculateNewVsModifiedCoverage()

	// Calculate test type coverage
	report.calculateTestTypeCoverage(testReports)

	// Generate insights and recommendations
	report.generateInsights(threshold)
	report.generateRecommendations(threshold)

	return report, nil
}

// calculateOverallMetrics calculates project-wide coverage metrics
func (r *HealthReport) calculateOverallMetrics(coverageReport *coverage.Report) {
	for _, fileCoverage := range coverageReport.FileCoverage {
		r.TotalFiles++
		r.TotalLines += fileCoverage.TotalLines
		r.TotalCoveredLines += fileCoverage.CoveredLines
		r.TotalUncoveredLines += fileCoverage.TotalLines - fileCoverage.CoveredLines
	}

	if r.TotalLines > 0 {
		r.OverallCoverage = float64(r.TotalCoveredLines) / float64(r.TotalLines) * 100
	}
}

// analyzeFileHealth analyzes health for a single file
func (r *HealthReport) analyzeFileHealth(
	filePath string,
	changedLines map[int]bool,
	diffResult *hunk.ParseResult,
	testReports []*TestCoverageReport,
	baselineReport *coverage.Report,
	threshold float64,
) *FileHealth {
	health := &FileHealth{
		FilePath:  filePath,
		IsNewFile: diffResult.IsNewFile(filePath),
	}

	// Aggregate current coverage for this file
	currentAggregated, _ := AggregateCoverage(testReports)
	fileCoverage := currentAggregated.GetCoverageForFile(filePath)
	if fileCoverage == nil {
		normalizedPath := coverage.NormalizePath(filePath)
		fileCoverage = currentAggregated.GetCoverageForFile(normalizedPath)
	}

	if fileCoverage != nil {
		health.TotalLines = fileCoverage.TotalLines
		health.CoveredLines = fileCoverage.CoveredLines
		health.UncoveredLines = fileCoverage.TotalLines - fileCoverage.CoveredLines
		if health.TotalLines > 0 {
			health.CoveragePercentage = float64(health.CoveredLines) / float64(health.TotalLines) * 100
		}
	}

	// Analyze changed lines specifically
	health.ChangedLines = len(changedLines)
	health.ChangedCoveredLines = 0
	health.ChangedUncoveredLines = 0

	// Track which test types cover changed lines
	health.ChangedLinesCoveredByUnit = 0
	health.ChangedLinesCoveredByAPI = 0
	health.ChangedLinesCoveredByFunctional = 0

	for lineNum := range changedLines {
		isCovered := false
		if fileCoverage != nil {
			isCovered = currentAggregated.IsLineCovered(filePath, lineNum)
			if !isCovered {
				normalizedPath := coverage.NormalizePath(filePath)
				isCovered = currentAggregated.IsLineCovered(normalizedPath, lineNum)
			}
		}

		if isCovered {
			health.ChangedCoveredLines++

			// Check which test types cover this line
			for _, testReport := range testReports {
				if testReport.CoverageReport == nil {
					continue
				}
				testFileCoverage := testReport.CoverageReport.GetCoverageForFile(filePath)
				if testFileCoverage == nil {
					normalizedPath := coverage.NormalizePath(filePath)
					testFileCoverage = testReport.CoverageReport.GetCoverageForFile(normalizedPath)
				}
				if testFileCoverage != nil && testFileCoverage.LineHits[lineNum] > 0 {
					switch testReport.TestType {
					case TestTypeUnit:
						health.ChangedLinesCoveredByUnit++
					case TestTypeAPI:
						health.ChangedLinesCoveredByAPI++
					case TestTypeFunctional:
						health.ChangedLinesCoveredByFunctional++
					}
				}
			}
		} else {
			health.ChangedUncoveredLines++
		}
	}

	if health.ChangedLines > 0 {
		health.ChangedCoveragePercentage = float64(health.ChangedCoveredLines) / float64(health.ChangedLines) * 100
	}

	// Calculate baseline comparison
	if baselineReport != nil && !health.IsNewFile {
		baselineFileCoverage := baselineReport.GetCoverageForFile(filePath)
		if baselineFileCoverage == nil {
			normalizedPath := coverage.NormalizePath(filePath)
			baselineFileCoverage = baselineReport.GetCoverageForFile(normalizedPath)
		}

		if baselineFileCoverage != nil {
			baselineCovered := 0
			for lineNum := range changedLines {
				if baselineReport.IsLineCovered(filePath, lineNum) || baselineReport.IsLineCovered(coverage.NormalizePath(filePath), lineNum) {
					baselineCovered++
				}
			}
			if len(changedLines) > 0 {
				health.BaselineCoveragePercentage = float64(baselineCovered) / float64(len(changedLines)) * 100
			}
			health.CoverageDelta = health.ChangedCoveragePercentage - health.BaselineCoveragePercentage
			health.HasRegression = health.CoverageDelta < 0 && health.ChangedCoveragePercentage < threshold
		}
	}

	// Calculate per-test-type coverage for changed lines
	if health.ChangedLines > 0 {
		health.UnitTestCoverage = float64(health.ChangedLinesCoveredByUnit) / float64(health.ChangedLines) * 100
		health.APITestCoverage = float64(health.ChangedLinesCoveredByAPI) / float64(health.ChangedLines) * 100
		health.FunctionalTestCoverage = float64(health.ChangedLinesCoveredByFunctional) / float64(health.ChangedLines) * 100
	}

	// Determine if file needs attention
	health.NeedsAttention = health.ChangedCoveragePercentage < threshold || health.HasRegression

	return health
}

// calculateNewVsModifiedCoverage calculates coverage for new vs modified files
func (r *HealthReport) calculateNewVsModifiedCoverage() {
	var newFilesCovered, newFilesTotal int
	var modifiedFilesCovered, modifiedFilesTotal int

	for _, fileHealth := range r.FileHealth {
		if fileHealth.IsNewFile {
			newFilesTotal += fileHealth.ChangedLines
			newFilesCovered += fileHealth.ChangedCoveredLines
		} else {
			modifiedFilesTotal += fileHealth.ChangedLines
			modifiedFilesCovered += fileHealth.ChangedCoveredLines
		}
	}

	if newFilesTotal > 0 {
		r.NewFilesCoverage = float64(newFilesCovered) / float64(newFilesTotal) * 100
	}
	if modifiedFilesTotal > 0 {
		r.ModifiedFilesCoverage = float64(modifiedFilesCovered) / float64(modifiedFilesTotal) * 100
	}
}

// calculateTestTypeCoverage calculates coverage per test type
func (r *HealthReport) calculateTestTypeCoverage(testReports []*TestCoverageReport) {
	var unitCovered, unitTotal int
	var apiCovered, apiTotal int
	var functionalCovered, functionalTotal int

	for _, testReport := range testReports {
		if testReport.CoverageReport == nil {
			continue
		}

		for _, fileCoverage := range testReport.CoverageReport.FileCoverage {
			switch testReport.TestType {
			case TestTypeUnit:
				unitTotal += fileCoverage.TotalLines
				unitCovered += fileCoverage.CoveredLines
			case TestTypeAPI:
				apiTotal += fileCoverage.TotalLines
				apiCovered += fileCoverage.CoveredLines
			case TestTypeFunctional:
				functionalTotal += fileCoverage.TotalLines
				functionalCovered += fileCoverage.CoveredLines
			}
		}
	}

	if unitTotal > 0 {
		r.UnitTestCoverage = float64(unitCovered) / float64(unitTotal) * 100
	}
	if apiTotal > 0 {
		r.APITestCoverage = float64(apiCovered) / float64(apiTotal) * 100
	}
	if functionalTotal > 0 {
		r.FunctionalTestCoverage = float64(functionalCovered) / float64(functionalTotal) * 100
	}
}

// generateInsights generates actionable insights about testing health
func (r *HealthReport) generateInsights(threshold float64) {
	// Count healthy and at-risk files
	for _, fileHealth := range r.FileHealth {
		if fileHealth.ChangedCoveragePercentage >= threshold {
			r.HealthyFiles++
		} else {
			r.AtRiskFiles++
		}
		if fileHealth.HasRegression {
			r.RegressingFiles++
		}
	}

	// Generate insights
	if r.ChangedCoverage >= threshold {
		r.Insights = append(r.Insights, Insight{
			Type:        "success",
			Category:    "coverage",
			Title:       "Coverage threshold met",
			Description: fmt.Sprintf("Overall change coverage is %.1f%%, meeting the %.1f%% threshold", r.ChangedCoverage, threshold),
			Severity:    "low",
		})
	} else {
		r.Insights = append(r.Insights, Insight{
			Type:        "warning",
			Category:    "coverage",
			Title:       "Coverage below threshold",
			Description: fmt.Sprintf("Overall change coverage is %.1f%%, below the %.1f%% threshold", r.ChangedCoverage, threshold),
			Severity:    "high",
		})
	}

	// Regression insights
	if r.RegressingFiles > 0 {
		r.Insights = append(r.Insights, Insight{
			Type:        "error",
			Category:    "regression",
			Title:       fmt.Sprintf("%d file(s) with coverage regression", r.RegressingFiles),
			Description: "Coverage has dropped below baseline for some files",
			Severity:    "critical",
		})
	}

	// Test type gap insights
	for _, fileHealth := range r.FileHealth {
		if fileHealth.ChangedUncoveredLines > 0 {
			// Check if there are test type gaps
			if fileHealth.ChangedLinesCoveredByUnit == 0 && fileHealth.ChangedLinesCoveredByAPI == 0 {
				r.Insights = append(r.Insights, Insight{
					Type:        "warning",
					Category:    "test-gap",
					Title:       "No test coverage for changed lines",
					Description: fmt.Sprintf("%s has %d uncovered changed lines", fileHealth.FilePath, fileHealth.ChangedUncoveredLines),
					File:        fileHealth.FilePath,
					Severity:    "high",
				})
			} else if fileHealth.ChangedLinesCoveredByUnit > 0 && fileHealth.ChangedLinesCoveredByAPI == 0 && fileHealth.ChangedLinesCoveredByFunctional == 0 {
				r.Insights = append(r.Insights, Insight{
					Type:        "info",
					Category:    "test-gap",
					Title:       "Unit tests only, consider integration tests",
					Description: fmt.Sprintf("%s is covered by unit tests but not API/functional tests", fileHealth.FilePath),
					File:        fileHealth.FilePath,
					Severity:    "medium",
				})
			}
		}
	}
}

// generateRecommendations generates actionable recommendations
func (r *HealthReport) generateRecommendations(threshold float64) {
	// Files that need attention
	var filesNeedingTests []string
	var filesWithRegression []string

	for filePath, fileHealth := range r.FileHealth {
		if fileHealth.NeedsAttention {
			if fileHealth.HasRegression {
				filesWithRegression = append(filesWithRegression, filePath)
			} else if fileHealth.ChangedCoveragePercentage < threshold {
				filesNeedingTests = append(filesNeedingTests, filePath)
			}
		}
	}

	// Generate recommendations
	if len(filesNeedingTests) > 0 {
		r.Recommendations = append(r.Recommendations, Recommendation{
			Priority:    "high",
			Category:    "add-tests",
			Title:       "Add tests for uncovered changes",
			Description: fmt.Sprintf("%d file(s) have changes below coverage threshold", len(filesNeedingTests)),
			Action:      "Add unit tests for uncovered changed lines",
			Files:       filesNeedingTests,
			TestType:    TestTypeUnit,
		})
	}

	if len(filesWithRegression) > 0 {
		r.Recommendations = append(r.Recommendations, Recommendation{
			Priority:    "critical",
			Category:    "fix-regression",
			Title:       "Fix coverage regression",
			Description: fmt.Sprintf("%d file(s) have coverage below baseline", len(filesWithRegression)),
			Action:      "Restore test coverage to baseline levels",
			Files:       filesWithRegression,
			TestType:    TestTypeUnit,
		})
	}
}
