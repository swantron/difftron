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
	
	// NewFileMetrics tracks coverage for new files only
	NewFileMetrics *FileTypeMetrics
	// ModifiedFileMetrics tracks coverage for modified files only
	ModifiedFileMetrics *FileTypeMetrics
}

// FileTypeMetrics tracks coverage metrics for a specific type of files (new or modified)
type FileTypeMetrics struct {
	TotalChangedLines int
	CoveredLines      int
	UncoveredLines    int
	CoveragePercentage float64
	FileCount         int
}

// FileResult contains analysis results for a single file
type FileResult struct {
	FilePath           string
	TotalChangedLines  int
	CoveredLines       int
	UncoveredLines     int
	CoveragePercentage float64
	// UncoveredLineNumbers lists the line numbers that are not covered
	UncoveredLineNumbers []int
	// CoveredLineNumbers lists the line numbers that are covered
	CoveredLineNumbers []int
	// IsNewFile indicates if this file is new (didn't exist in base)
	IsNewFile bool
	// BaselineCoveragePercentage is the coverage percentage before changes (for modified files)
	// This helps identify if coverage actually dropped or if we're just seeing untested code for the first time
	BaselineCoveragePercentage float64
}

// Analyze compares git diff hunks with coverage data
func Analyze(diffResult *hunk.ParseResult, coverageReport *coverage.Report) (*AnalysisResult, error) {
	return AnalyzeWithBaseline(diffResult, coverageReport, nil)
}

// AnalyzeWithBaseline compares git diff hunks with coverage data, accounting for baseline coverage
// baselineReport can be nil if baseline coverage is not available
func AnalyzeWithBaseline(diffResult *hunk.ParseResult, coverageReport *coverage.Report, baselineReport *coverage.Report) (*AnalysisResult, error) {
	if diffResult == nil {
		return nil, fmt.Errorf("diff result cannot be nil")
	}
	if coverageReport == nil {
		return nil, fmt.Errorf("coverage report cannot be nil")
	}

	result := &AnalysisResult{
		FileResults:        make(map[string]*FileResult),
		NewFileMetrics:     &FileTypeMetrics{},
		ModifiedFileMetrics: &FileTypeMetrics{},
	}

	// Process each changed file
	for filePath, changedLines := range diffResult.ChangedLines {
		isNewFile := diffResult.IsNewFile(filePath)
		fileResult := analyzeFile(filePath, changedLines, coverageReport, baselineReport, isNewFile)
		result.FileResults[filePath] = fileResult

		// Update overall metrics
		result.TotalChangedLines += fileResult.TotalChangedLines
		result.CoveredLines += fileResult.CoveredLines
		result.UncoveredLines += fileResult.UncoveredLines

		// Update type-specific metrics
		if isNewFile {
			result.NewFileMetrics.TotalChangedLines += fileResult.TotalChangedLines
			result.NewFileMetrics.CoveredLines += fileResult.CoveredLines
			result.NewFileMetrics.UncoveredLines += fileResult.UncoveredLines
			result.NewFileMetrics.FileCount++
		} else {
			result.ModifiedFileMetrics.TotalChangedLines += fileResult.TotalChangedLines
			result.ModifiedFileMetrics.CoveredLines += fileResult.CoveredLines
			result.ModifiedFileMetrics.UncoveredLines += fileResult.UncoveredLines
			result.ModifiedFileMetrics.FileCount++
		}
	}

	// Calculate overall coverage percentage
	if result.TotalChangedLines > 0 {
		result.CoveragePercentage = float64(result.CoveredLines) / float64(result.TotalChangedLines) * 100
	}

	// Calculate type-specific coverage percentages
	if result.NewFileMetrics.TotalChangedLines > 0 {
		result.NewFileMetrics.CoveragePercentage = float64(result.NewFileMetrics.CoveredLines) / float64(result.NewFileMetrics.TotalChangedLines) * 100
	}
	if result.ModifiedFileMetrics.TotalChangedLines > 0 {
		result.ModifiedFileMetrics.CoveragePercentage = float64(result.ModifiedFileMetrics.CoveredLines) / float64(result.ModifiedFileMetrics.TotalChangedLines) * 100
	}

	return result, nil
}

// analyzeFile analyzes coverage for a single file
func analyzeFile(filePath string, changedLines map[int]bool, coverageReport *coverage.Report, baselineReport *coverage.Report, isNewFile bool) *FileResult {
	fileResult := &FileResult{
		FilePath:             filePath,
		UncoveredLineNumbers: make([]int, 0),
		CoveredLineNumbers:   make([]int, 0),
		IsNewFile:            isNewFile,
	}

	// Try multiple path variations to match coverage data
	normalizedPath := coverage.NormalizePath(filePath)
	var fileCoverage *coverage.CoverageData
	var baselineFileCoverage *coverage.CoverageData

	// Try exact match first
	fileCoverage = coverageReport.GetCoverageForFile(filePath)
	if fileCoverage == nil {
		// Try normalized path
		fileCoverage = coverageReport.GetCoverageForFile(normalizedPath)
	}

	// Get baseline coverage for modified files
	if !isNewFile && baselineReport != nil {
		baselineFileCoverage = baselineReport.GetCoverageForFile(filePath)
		if baselineFileCoverage == nil {
			baselineFileCoverage = baselineReport.GetCoverageForFile(normalizedPath)
		}
		
		// Calculate baseline coverage percentage for the changed lines
		if baselineFileCoverage != nil {
			baselineCovered := 0
			baselineTotal := len(changedLines)
			for lineNum := range changedLines {
				if baselineReport.IsLineCovered(filePath, lineNum) || baselineReport.IsLineCovered(normalizedPath, lineNum) {
					baselineCovered++
				}
			}
			if baselineTotal > 0 {
				fileResult.BaselineCoveragePercentage = float64(baselineCovered) / float64(baselineTotal) * 100
			}
		}
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
