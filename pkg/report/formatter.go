package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/swantron/difftron/internal/analyzer"
)

// AnalysisReport represents the JSON output structure for analyze command
type AnalysisReport struct {
	TotalChangedLines  int                    `json:"total_changed_lines"`
	CoveredLines       int                    `json:"covered_lines"`
	UncoveredLines     int                    `json:"uncovered_lines"`
	CoveragePercentage float64                `json:"coverage_percentage"`
	MeetsThreshold     bool                   `json:"meets_threshold"`
	Threshold          float64                `json:"threshold,omitempty"`
	Files              map[string]*FileReport `json:"files"`
	NewFiles           *FileTypeReport        `json:"new_files,omitempty"`
	ModifiedFiles      *FileTypeReport        `json:"modified_files,omitempty"`
}

// FileReport represents file-level analysis results
type FileReport struct {
	FilePath             string  `json:"file_path"`
	CoveragePercentage   float64 `json:"coverage_percentage"`
	CoveredLines         int     `json:"covered_lines"`
	UncoveredLines       int     `json:"uncovered_lines"`
	TotalChangedLines    int     `json:"total_changed_lines"`
	UncoveredLineNumbers []int   `json:"uncovered_line_numbers"`
	CoveredLineNumbers   []int   `json:"covered_line_numbers,omitempty"`
	IsNewFile            bool    `json:"is_new_file"`
	BaselineCoverage     float64 `json:"baseline_coverage,omitempty"`
}

// FileTypeReport represents metrics for new or modified files
type FileTypeReport struct {
	FileCount          int     `json:"file_count"`
	TotalChangedLines  int     `json:"total_changed_lines"`
	CoveredLines       int     `json:"covered_lines"`
	UncoveredLines     int     `json:"uncovered_lines"`
	CoveragePercentage float64 `json:"coverage_percentage"`
}

// ToJSON converts an AnalysisResult to JSON format
func ToJSON(result *analyzer.AnalysisResult, threshold float64) ([]byte, error) {
	report := &AnalysisReport{
		TotalChangedLines:  result.TotalChangedLines,
		CoveredLines:       result.CoveredLines,
		UncoveredLines:     result.UncoveredLines,
		CoveragePercentage: result.CoveragePercentage,
		MeetsThreshold:     result.MeetsThreshold(threshold),
		Threshold:          threshold,
		Files:              make(map[string]*FileReport),
	}

	// Convert file results
	for filePath, fileResult := range result.FileResults {
		report.Files[filePath] = &FileReport{
			FilePath:             fileResult.FilePath,
			CoveragePercentage:   fileResult.CoveragePercentage,
			CoveredLines:         fileResult.CoveredLines,
			UncoveredLines:       fileResult.UncoveredLines,
			TotalChangedLines:    fileResult.TotalChangedLines,
			UncoveredLineNumbers: fileResult.UncoveredLineNumbers,
			CoveredLineNumbers:   fileResult.CoveredLineNumbers,
			IsNewFile:            fileResult.IsNewFile,
			BaselineCoverage:     fileResult.BaselineCoveragePercentage,
		}
	}

	// Add new/modified file metrics if available
	if result.NewFileMetrics != nil && result.NewFileMetrics.FileCount > 0 {
		report.NewFiles = &FileTypeReport{
			FileCount:          result.NewFileMetrics.FileCount,
			TotalChangedLines:  result.NewFileMetrics.TotalChangedLines,
			CoveredLines:       result.NewFileMetrics.CoveredLines,
			UncoveredLines:     result.NewFileMetrics.UncoveredLines,
			CoveragePercentage: result.NewFileMetrics.CoveragePercentage,
		}
	}

	if result.ModifiedFileMetrics != nil && result.ModifiedFileMetrics.FileCount > 0 {
		report.ModifiedFiles = &FileTypeReport{
			FileCount:          result.ModifiedFileMetrics.FileCount,
			TotalChangedLines:  result.ModifiedFileMetrics.TotalChangedLines,
			CoveredLines:       result.ModifiedFileMetrics.CoveredLines,
			UncoveredLines:     result.ModifiedFileMetrics.UncoveredLines,
			CoveragePercentage: result.ModifiedFileMetrics.CoveragePercentage,
		}
	}

	return json.MarshalIndent(report, "", "  ")
}

// ToMarkdown converts an AnalysisResult to Markdown format
func ToMarkdown(result *analyzer.AnalysisResult, threshold float64) string {
	var sb strings.Builder

	sb.WriteString("# Coverage Analysis Report\n\n")

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Overall Coverage**: %.1f%%\n", result.CoveragePercentage))
	sb.WriteString(fmt.Sprintf("- **Changed Lines**: %d\n", result.TotalChangedLines))
	sb.WriteString(fmt.Sprintf("- **Covered Lines**: %d\n", result.CoveredLines))
	sb.WriteString(fmt.Sprintf("- **Uncovered Lines**: %d\n", result.UncoveredLines))
	sb.WriteString(fmt.Sprintf("- **Threshold**: %.1f%%\n", threshold))

	meetsThreshold := result.MeetsThreshold(threshold)
	status := "❌ FAIL"
	if meetsThreshold {
		status = "✅ PASS"
	}
	sb.WriteString(fmt.Sprintf("- **Status**: %s\n\n", status))

	// New vs Modified breakdown
	if result.NewFileMetrics != nil && result.NewFileMetrics.FileCount > 0 {
		sb.WriteString("### New Files\n\n")
		sb.WriteString(fmt.Sprintf("- **Files**: %d\n", result.NewFileMetrics.FileCount))
		sb.WriteString(fmt.Sprintf("- **Coverage**: %.1f%%\n", result.NewFileMetrics.CoveragePercentage))
		sb.WriteString(fmt.Sprintf("- **Changed Lines**: %d\n", result.NewFileMetrics.TotalChangedLines))
		sb.WriteString(fmt.Sprintf("- **Covered**: %d | **Uncovered**: %d\n\n",
			result.NewFileMetrics.CoveredLines, result.NewFileMetrics.UncoveredLines))
	}

	if result.ModifiedFileMetrics != nil && result.ModifiedFileMetrics.FileCount > 0 {
		sb.WriteString("### Modified Files\n\n")
		sb.WriteString(fmt.Sprintf("- **Files**: %d\n", result.ModifiedFileMetrics.FileCount))
		sb.WriteString(fmt.Sprintf("- **Coverage**: %.1f%%\n", result.ModifiedFileMetrics.CoveragePercentage))
		sb.WriteString(fmt.Sprintf("- **Changed Lines**: %d\n", result.ModifiedFileMetrics.TotalChangedLines))
		sb.WriteString(fmt.Sprintf("- **Covered**: %d | **Uncovered**: %d\n\n",
			result.ModifiedFileMetrics.CoveredLines, result.ModifiedFileMetrics.UncoveredLines))
	}

	// Per-file results
	if len(result.FileResults) > 0 {
		sb.WriteString("## File Details\n\n")
		sb.WriteString("| File | Status | Coverage | Changed | Covered | Uncovered |\n")
		sb.WriteString("|------|--------|----------|---------|---------|-----------|\n")

		for filePath, fileResult := range result.FileResults {
			fileStatus := "✅"
			if fileResult.CoveragePercentage < threshold {
				fileStatus = "❌"
			}
			if fileResult.IsNewFile {
				fileStatus += " NEW"
			}

			sb.WriteString(fmt.Sprintf("| `%s` | %s | %.1f%% | %d | %d | %d |\n",
				filePath, fileStatus, fileResult.CoveragePercentage,
				fileResult.TotalChangedLines, fileResult.CoveredLines, fileResult.UncoveredLines))

			// Add uncovered lines if any
			if len(fileResult.UncoveredLineNumbers) > 0 {
				sb.WriteString(fmt.Sprintf("  - Uncovered lines: %v\n", fileResult.UncoveredLineNumbers))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
