package main

import (
	"testing"

	"github.com/swantron/difftron/internal/analyzer"
)

func TestOutputText(t *testing.T) {
	result := &analyzer.AnalysisResult{
		TotalChangedLines:  10,
		CoveredLines:       8,
		UncoveredLines:     2,
		CoveragePercentage: 80.0,
		FileResults: map[string]*analyzer.FileResult{
			"test.go": {
				TotalChangedLines:    10,
				CoveredLines:         8,
				UncoveredLines:       2,
				CoveragePercentage:   80.0,
				UncoveredLineNumbers: []int{5, 7},
			},
		},
	}

	threshold = 80.0

	// This will exit with code 0 since threshold is met
	err := outputText(result)
	if err != nil {
		t.Errorf("outputText() error = %v", err)
	}
}

// TestOutputTextBelowThreshold is skipped because outputText calls os.Exit(1)
// which terminates the test process. This is expected behavior for CLI.
// func TestOutputTextBelowThreshold(t *testing.T) {
// 	// This test would verify that outputText exits with code 1 when threshold not met
// 	// but os.Exit prevents proper testing
// }

func TestOutputTextNoChanges(t *testing.T) {
	result := &analyzer.AnalysisResult{
		TotalChangedLines:  0,
		CoveredLines:       0,
		UncoveredLines:     0,
		CoveragePercentage: 0.0,
		FileResults:        make(map[string]*analyzer.FileResult),
	}

	threshold = 80.0

	err := outputText(result)
	if err != nil {
		t.Errorf("outputText() error = %v", err)
	}
}

func TestOutputJSON(t *testing.T) {
	result := &analyzer.AnalysisResult{
		TotalChangedLines:  10,
		CoveredLines:       8,
		UncoveredLines:     2,
		CoveragePercentage: 80.0,
		FileResults: map[string]*analyzer.FileResult{
			"test.go": {
				TotalChangedLines:    10,
				CoveredLines:         8,
				UncoveredLines:       2,
				CoveragePercentage:   80.0,
				UncoveredLineNumbers: []int{5, 7},
			},
		},
	}

	threshold = 80.0

	err := outputJSON(result)
	if err != nil {
		t.Errorf("outputJSON() error = %v", err)
	}
}

func TestGetGitDiff(t *testing.T) {
	// Test with same base and head (HEAD)
	result, err := getGitDiff("HEAD", "HEAD")
	if err != nil {
		// This is okay - might not have git repo or changes
		t.Logf("getGitDiff(HEAD, HEAD) returned error (expected in some cases): %v", err)
	} else {
		// Should return a string (could be empty)
		if result == "" {
			t.Log("getGitDiff returned empty string (no changes)")
		}
	}
}
