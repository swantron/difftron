package health

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatHealthReport formats a health report for output
type FormatHealthReport struct {
	Summary       SummarySection       `json:"summary"`
	TestTypes     TestTypeSection      `json:"test_types"`
	Changes       ChangesSection       `json:"changes"`
	Files         []FileSection        `json:"files"`
	Insights      []InsightSection     `json:"insights"`
	Recommendations []RecommendationSection `json:"recommendations"`
}

type SummarySection struct {
	OverallCoverage      float64 `json:"overall_coverage"`
	ChangedCoverage      float64 `json:"changed_coverage"`
	TotalFiles           int     `json:"total_files"`
	ChangedFiles         int     `json:"changed_files"`
	HealthyFiles         int     `json:"healthy_files"`
	AtRiskFiles          int     `json:"at_risk_files"`
	RegressingFiles      int     `json:"regressing_files"`
	NewFilesCount        int     `json:"new_files_count"`
	ModifiedFilesCount   int     `json:"modified_files_count"`
	NewFilesCoverage     float64 `json:"new_files_coverage"`
	ModifiedFilesCoverage float64 `json:"modified_files_coverage"`
}

type TestTypeSection struct {
	UnitTestCoverage      float64 `json:"unit_test_coverage"`
	APITestCoverage       float64 `json:"api_test_coverage"`
	FunctionalTestCoverage float64 `json:"functional_test_coverage"`
}

type ChangesSection struct {
	TotalChangedLines     int     `json:"total_changed_lines"`
	CoveredLines          int     `json:"covered_lines"`
	UncoveredLines        int     `json:"uncovered_lines"`
	CoveragePercentage   float64 `json:"coverage_percentage"`
}

type FileSection struct {
	FilePath              string   `json:"file_path"`
	IsNewFile             bool     `json:"is_new_file"`
	OverallCoverage       float64  `json:"overall_coverage"`
	ChangedCoverage       float64  `json:"changed_coverage"`
	BaselineCoverage      float64  `json:"baseline_coverage"`
	CoverageDelta         float64  `json:"coverage_delta"`
	ChangedLines          int      `json:"changed_lines"`
	CoveredLines          int      `json:"covered_lines"`
	UncoveredLines        int      `json:"uncovered_lines"`
	UnitTestCoverage      float64  `json:"unit_test_coverage"`
	APITestCoverage       float64  `json:"api_test_coverage"`
	FunctionalTestCoverage float64 `json:"functional_test_coverage"`
	Status                string   `json:"status"` // "healthy", "at_risk", "regressing"
	UncoveredLineNumbers  []int    `json:"uncovered_line_numbers"`
}

type InsightSection struct {
	Type        string `json:"type"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	File        string `json:"file,omitempty"`
	Severity    string `json:"severity"`
}

type RecommendationSection struct {
	Priority    string   `json:"priority"`
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Action      string   `json:"action"`
	Files       []string `json:"files"`
	TestType    string   `json:"test_type"`
}

// ToJSON converts a HealthReport to JSON format
func (r *HealthReport) ToJSON() ([]byte, error) {
	formatted := r.toFormatted()
	return json.MarshalIndent(formatted, "", "  ")
}

// ToMarkdown converts a HealthReport to Markdown format
func (r *HealthReport) ToMarkdown() string {
	var sb strings.Builder

	// Header
	sb.WriteString("# Testing Health Report\n\n")

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Overall Coverage:** %.1f%%\n", r.OverallCoverage))
	sb.WriteString(fmt.Sprintf("**Changed Coverage:** %.1f%%\n", r.ChangedCoverage))
	sb.WriteString(fmt.Sprintf("**Total Files:** %d | **Changed Files:** %d\n", r.TotalFiles, r.ChangedFiles))
	sb.WriteString(fmt.Sprintf("**Healthy:** %d | **At Risk:** %d | **Regressing:** %d\n\n", r.HealthyFiles, r.AtRiskFiles, r.RegressingFiles))

	// Test Type Breakdown
	sb.WriteString("## Test Type Coverage\n\n")
	sb.WriteString("| Test Type | Coverage |\n")
	sb.WriteString("|-----------|----------|\n")
	sb.WriteString(fmt.Sprintf("| Unit Tests | %.1f%% |\n", r.UnitTestCoverage))
	sb.WriteString(fmt.Sprintf("| API Tests | %.1f%% |\n", r.APITestCoverage))
	sb.WriteString(fmt.Sprintf("| Functional Tests | %.1f%% |\n", r.FunctionalTestCoverage))
	sb.WriteString("\n")

	// Changes Breakdown
	sb.WriteString("## Changes Breakdown\n\n")
	sb.WriteString(fmt.Sprintf("- **New Files:** %d (%.1f%% coverage)\n", r.NewFilesCount, r.NewFilesCoverage))
	sb.WriteString(fmt.Sprintf("- **Modified Files:** %d (%.1f%% coverage)\n", r.ModifiedFilesCount, r.ModifiedFilesCoverage))
	sb.WriteString(fmt.Sprintf("- **Total Changed Lines:** %d\n", r.ChangedLines))
	sb.WriteString(fmt.Sprintf("- **Covered Lines:** %d\n", r.ChangedCoveredLines))
	sb.WriteString(fmt.Sprintf("- **Uncovered Lines:** %d\n\n", r.ChangedUncoveredLines))

	// File Details
	if len(r.FileHealth) > 0 {
		sb.WriteString("## File Health\n\n")
		sb.WriteString("| File | Status | Changed Coverage | Baseline | Delta | Test Types |\n")
		sb.WriteString("|------|--------|-----------------|----------|-------|------------|\n")

		for filePath, fileHealth := range r.FileHealth {
			status := "✅ Healthy"
			if fileHealth.HasRegression {
				status = "🔴 Regressing"
			} else if fileHealth.NeedsAttention {
				status = "⚠️ At Risk"
			}

			testTypes := []string{}
			if fileHealth.ChangedLinesCoveredByUnit > 0 {
				testTypes = append(testTypes, "Unit")
			}
			if fileHealth.ChangedLinesCoveredByAPI > 0 {
				testTypes = append(testTypes, "API")
			}
			if fileHealth.ChangedLinesCoveredByFunctional > 0 {
				testTypes = append(testTypes, "Functional")
			}
			testTypesStr := strings.Join(testTypes, ", ")
			if testTypesStr == "" {
				testTypesStr = "None"
			}

			deltaStr := fmt.Sprintf("%+.1f%%", fileHealth.CoverageDelta)
			if fileHealth.IsNewFile {
				deltaStr = "N/A"
			}

			sb.WriteString(fmt.Sprintf("| `%s` | %s | %.1f%% | %.1f%% | %s | %s |\n",
				filePath, status, fileHealth.ChangedCoveragePercentage,
				fileHealth.BaselineCoveragePercentage, deltaStr, testTypesStr))
		}
		sb.WriteString("\n")
	}

	// Insights
	if len(r.Insights) > 0 {
		sb.WriteString("## Insights\n\n")
		for _, insight := range r.Insights {
			icon := "ℹ️"
			switch insight.Type {
			case "success":
				icon = "✅"
			case "warning":
				icon = "⚠️"
			case "error":
				icon = "🔴"
			}

			sb.WriteString(fmt.Sprintf("### %s %s\n\n", icon, insight.Title))
			sb.WriteString(fmt.Sprintf("%s\n\n", insight.Description))
			if insight.File != "" {
				sb.WriteString(fmt.Sprintf("**File:** `%s`\n\n", insight.File))
			}
		}
	}

	// Recommendations
	if len(r.Recommendations) > 0 {
		sb.WriteString("## Recommendations\n\n")
		for _, rec := range r.Recommendations {
			priorityIcon := "📌"
			switch rec.Priority {
			case "critical":
				priorityIcon = "🔴"
			case "high":
				priorityIcon = "⚠️"
			case "medium":
				priorityIcon = "📋"
			}

			sb.WriteString(fmt.Sprintf("### %s %s (%s priority)\n\n", priorityIcon, rec.Title, rec.Priority))
			sb.WriteString(fmt.Sprintf("%s\n\n", rec.Description))
			sb.WriteString(fmt.Sprintf("**Action:** %s\n\n", rec.Action))
			if len(rec.Files) > 0 {
				sb.WriteString("**Files needing attention:**\n")
				for _, file := range rec.Files {
					sb.WriteString(fmt.Sprintf("- `%s`\n", file))
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// ToStructuredText converts to a structured text format for AI agents
func (r *HealthReport) ToStructuredText() string {
	var sb strings.Builder

	sb.WriteString("TESTING HEALTH REPORT\n")
	sb.WriteString("=====================\n\n")

	// Executive Summary
	sb.WriteString("EXECUTIVE SUMMARY\n")
	sb.WriteString("-----------------\n")
	sb.WriteString(fmt.Sprintf("Overall Project Coverage: %.1f%%\n", r.OverallCoverage))
	sb.WriteString(fmt.Sprintf("Changed Code Coverage: %.1f%%\n", r.ChangedCoverage))
	sb.WriteString(fmt.Sprintf("Files Analyzed: %d changed files out of %d total files\n", r.ChangedFiles, r.TotalFiles))
	sb.WriteString(fmt.Sprintf("Health Status: %d healthy, %d at risk, %d regressing\n\n", r.HealthyFiles, r.AtRiskFiles, r.RegressingFiles))

	// Test Type Analysis
	sb.WriteString("TEST TYPE ANALYSIS\n")
	sb.WriteString("-------------------\n")
	sb.WriteString(fmt.Sprintf("Unit Tests: %.1f%% coverage\n", r.UnitTestCoverage))
	sb.WriteString(fmt.Sprintf("API Tests: %.1f%% coverage\n", r.APITestCoverage))
	sb.WriteString(fmt.Sprintf("Functional Tests: %.1f%% coverage\n\n", r.FunctionalTestCoverage))

	// File-by-File Analysis
	sb.WriteString("FILE-BY-FILE ANALYSIS\n")
	sb.WriteString("---------------------\n")
	for filePath, fileHealth := range r.FileHealth {
		sb.WriteString(fmt.Sprintf("\nFile: %s\n", filePath))
		if fileHealth.IsNewFile {
			sb.WriteString("  Type: NEW FILE\n")
		} else {
			sb.WriteString("  Type: MODIFIED FILE\n")
		}
		sb.WriteString(fmt.Sprintf("  Changed Lines Coverage: %.1f%% (%d/%d lines covered)\n",
			fileHealth.ChangedCoveragePercentage, fileHealth.ChangedCoveredLines, fileHealth.ChangedLines))
		if !fileHealth.IsNewFile {
			sb.WriteString(fmt.Sprintf("  Baseline Coverage: %.1f%%\n", fileHealth.BaselineCoveragePercentage))
			sb.WriteString(fmt.Sprintf("  Coverage Delta: %+.1f%%\n", fileHealth.CoverageDelta))
		}
		sb.WriteString(fmt.Sprintf("  Test Types Covering Changes: Unit=%d, API=%d, Functional=%d\n",
			fileHealth.ChangedLinesCoveredByUnit, fileHealth.ChangedLinesCoveredByAPI, fileHealth.ChangedLinesCoveredByFunctional))
		if fileHealth.UncoveredLines > 0 {
			sb.WriteString(fmt.Sprintf("  Uncovered Lines: %d\n", fileHealth.UncoveredLines))
		}
		if fileHealth.HasRegression {
			sb.WriteString("  STATUS: REGRESSION - Coverage dropped below baseline\n")
		} else if fileHealth.NeedsAttention {
			sb.WriteString("  STATUS: AT RISK - Coverage below threshold\n")
		} else {
			sb.WriteString("  STATUS: HEALTHY\n")
		}
	}

	// Actionable Insights
	if len(r.Insights) > 0 {
		sb.WriteString("\n\nACTIONABLE INSIGHTS\n")
		sb.WriteString("-------------------\n")
		for _, insight := range r.Insights {
			sb.WriteString(fmt.Sprintf("\n[%s] %s\n", strings.ToUpper(insight.Severity), insight.Title))
			sb.WriteString(fmt.Sprintf("  %s\n", insight.Description))
			if insight.File != "" {
				sb.WriteString(fmt.Sprintf("  File: %s\n", insight.File))
			}
		}
	}

	// Recommendations
	if len(r.Recommendations) > 0 {
		sb.WriteString("\n\nRECOMMENDATIONS\n")
		sb.WriteString("---------------\n")
		for _, rec := range r.Recommendations {
			sb.WriteString(fmt.Sprintf("\n[%s PRIORITY] %s\n", strings.ToUpper(rec.Priority), rec.Title))
			sb.WriteString(fmt.Sprintf("  %s\n", rec.Description))
			sb.WriteString(fmt.Sprintf("  Action: %s\n", rec.Action))
			if len(rec.Files) > 0 {
				sb.WriteString("  Files:\n")
				for _, file := range rec.Files {
					sb.WriteString(fmt.Sprintf("    - %s\n", file))
				}
			}
		}
	}

	return sb.String()
}

// toFormatted converts HealthReport to FormatHealthReport
func (r *HealthReport) toFormatted() *FormatHealthReport {
	formatted := &FormatHealthReport{
		Summary: SummarySection{
			OverallCoverage:      r.OverallCoverage,
			ChangedCoverage:      r.ChangedCoverage,
			TotalFiles:           r.TotalFiles,
			ChangedFiles:         r.ChangedFiles,
			HealthyFiles:         r.HealthyFiles,
			AtRiskFiles:          r.AtRiskFiles,
			RegressingFiles:      r.RegressingFiles,
			NewFilesCount:        r.NewFilesCount,
			ModifiedFilesCount:   r.ModifiedFilesCount,
			NewFilesCoverage:     r.NewFilesCoverage,
			ModifiedFilesCoverage: r.ModifiedFilesCoverage,
		},
		TestTypes: TestTypeSection{
			UnitTestCoverage:      r.UnitTestCoverage,
			APITestCoverage:       r.APITestCoverage,
			FunctionalTestCoverage: r.FunctionalTestCoverage,
		},
		Changes: ChangesSection{
			TotalChangedLines:   r.ChangedLines,
			CoveredLines:        r.ChangedCoveredLines,
			UncoveredLines:      r.ChangedUncoveredLines,
			CoveragePercentage:  r.ChangedCoverage,
		},
		Files:         make([]FileSection, 0),
		Insights:      make([]InsightSection, 0),
		Recommendations: make([]RecommendationSection, 0),
	}

	// Convert file health
	for filePath, fileHealth := range r.FileHealth {
		status := "healthy"
		if fileHealth.HasRegression {
			status = "regressing"
		} else if fileHealth.NeedsAttention {
			status = "at_risk"
		}

		formatted.Files = append(formatted.Files, FileSection{
			FilePath:              filePath,
			IsNewFile:             fileHealth.IsNewFile,
			OverallCoverage:       fileHealth.CoveragePercentage,
			ChangedCoverage:       fileHealth.ChangedCoveragePercentage,
			BaselineCoverage:      fileHealth.BaselineCoveragePercentage,
			CoverageDelta:         fileHealth.CoverageDelta,
			ChangedLines:          fileHealth.ChangedLines,
			CoveredLines:          fileHealth.ChangedCoveredLines,
			UncoveredLines:        fileHealth.ChangedUncoveredLines,
			UnitTestCoverage:      fileHealth.UnitTestCoverage,
			APITestCoverage:       fileHealth.APITestCoverage,
			FunctionalTestCoverage: fileHealth.FunctionalTestCoverage,
			Status:                status,
		})
	}

	// Convert insights
	for _, insight := range r.Insights {
		formatted.Insights = append(formatted.Insights, InsightSection{
			Type:        insight.Type,
			Category:    insight.Category,
			Title:       insight.Title,
			Description: insight.Description,
			File:        insight.File,
			Severity:    insight.Severity,
		})
	}

	// Convert recommendations
	for _, rec := range r.Recommendations {
		formatted.Recommendations = append(formatted.Recommendations, RecommendationSection{
			Priority:    rec.Priority,
			Category:    rec.Category,
			Title:       rec.Title,
			Description: rec.Description,
			Action:      rec.Action,
			Files:       rec.Files,
			TestType:    string(rec.TestType),
		})
	}

	return formatted
}
