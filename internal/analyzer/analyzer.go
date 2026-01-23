package analyzer

import (
	"fmt"

	"github.com/swantron/difftron/internal/coverage"
	"github.com/swantron/difftron/internal/hunk"
)

// AnalysisResult contains the results of analyzing a diff against coverage
type AnalysisResult struct {
	// TotalChangedLines is the total number of lines changed in the diff
	TotalChangedLines int
	// CoveredLines is the number of changed lines that are covered
	CoveredLines int
	// UncoveredLines is the number of changed lines that are not covered
	UncoveredLines int
	// CoveragePercentage is the percentage of changed lines that are covered
	CoveragePercentage float64
	// FileResults contains per-file analysis results
	FileResults map[string]*FileResult
}

// FileResult contains analysis results for a single file
type FileResult struct {
	FilePath          string
	TotalChangedLines int
	CoveredLines      int
	UncoveredLines    int
	CoveragePercentage float64
	// UncoveredLineNumbers lists the line numbers that are not covered
	UncoveredLineNumbers []int
	// CoveredLineNumbers lists the line numbers that are covered
	CoveredLineNumbers []int
}

// Analyze compares git diff hunks with coverage data
func Analyze(diffResult *hunk.ParseResult, coverageReport *coverage.Report) (*AnalysisResult, error) {
	if diffResult == nil {
		return nil, fmt.Errorf("diff result cannot be nil")
	}
	if coverageReport == nil {
		return nil, fmt.Errorf("coverage report cannot be nil")
	}

	result := &AnalysisResult{
		FileResults: make(map[string]*FileResult),
	}

	// Process each changed file
	for filePath, changedLines := range diffResult.ChangedLines {
		fileResult := analyzeFile(filePath, changedLines, coverageReport)
		result.FileResults[filePath] = fileResult

		result.TotalChangedLines += fileResult.TotalChangedLines
		result.CoveredLines += fileResult.CoveredLines
		result.UncoveredLines += fileResult.UncoveredLines
	}

	// Calculate overall coverage percentage
	if result.TotalChangedLines > 0 {
		result.CoveragePercentage = float64(result.CoveredLines) / float64(result.TotalChangedLines) * 100
	}

	return result, nil
}

// analyzeFile analyzes coverage for a single file
func analyzeFile(filePath string, changedLines map[int]bool, coverageReport *coverage.Report) *FileResult {
	fileResult := &FileResult{
		FilePath:            filePath,
		UncoveredLineNumbers: make([]int, 0),
		CoveredLineNumbers:  make([]int, 0),
	}

	// Try multiple path variations to match coverage data
	normalizedPath := coverage.NormalizePath(filePath)
	var fileCoverage *coverage.CoverageData

	// Try exact match first
	fileCoverage = coverageReport.GetCoverageForFile(filePath)
	if fileCoverage == nil {
		// Try normalized path
		fileCoverage = coverageReport.GetCoverageForFile(normalizedPath)
	}

	// Process each changed line
	for lineNum := range changedLines {
		fileResult.TotalChangedLines++

		var isCovered bool
		if fileCoverage != nil {
			isCovered = coverageReport.IsLineCovered(filePath, lineNum)
			if !isCovered {
				// Try normalized path
				isCovered = coverageReport.IsLineCovered(normalizedPath, lineNum)
			}
		}

		if isCovered {
			fileResult.CoveredLines++
			fileResult.CoveredLineNumbers = append(fileResult.CoveredLineNumbers, lineNum)
		} else {
			fileResult.UncoveredLines++
			fileResult.UncoveredLineNumbers = append(fileResult.UncoveredLineNumbers, lineNum)
		}
	}

	// Calculate file-level coverage percentage
	if fileResult.TotalChangedLines > 0 {
		fileResult.CoveragePercentage = float64(fileResult.CoveredLines) / float64(fileResult.TotalChangedLines) * 100
	}

	return fileResult
}

// MeetsThreshold checks if the analysis result meets the specified coverage threshold
func (r *AnalysisResult) MeetsThreshold(threshold float64) bool {
	return r.CoveragePercentage >= threshold
}

// HasUncoveredLines returns true if there are any uncovered lines
func (r *AnalysisResult) HasUncoveredLines() bool {
	return r.UncoveredLines > 0
}
