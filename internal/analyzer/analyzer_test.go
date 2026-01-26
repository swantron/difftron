package analyzer

import (
	"os"
	"testing"

	"github.com/swantron/difftron/internal/coverage"
	"github.com/swantron/difftron/internal/hunk"
)

func TestAnalyze(t *testing.T) {
	// Create test diff
	diffOutput := `diff --git a/file.go b/file.go
index 123..456 100644
--- a/file.go
+++ b/file.go
@@ -5,3 +5,5 @@ func main() {
 	fmt.Println("hello")
+	fmt.Println("new line 1")
+	fmt.Println("new line 2")
 	fmt.Println("world")
`

	// Create test LCOV file
	lcovContent := `TN:
SF:file.go
DA:6,5
DA:7,0
DA:8,3
end_of_record
`

	// Parse diff
	diffResult, err := hunk.ParseGitDiff(diffOutput)
	if err != nil {
		t.Fatalf("failed to parse diff: %v", err)
	}

	// Parse coverage
	tmpfile, err := os.CreateTemp("", "test-*.info")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(lcovContent)); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}
	tmpfile.Close()

	coverageReport, err := coverage.ParseLCOV(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to parse coverage: %v", err)
	}

	// Analyze
	result, err := Analyze(diffResult, coverageReport)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check results
	if result.TotalChangedLines != 2 {
		t.Errorf("expected 2 total changed lines, got %d", result.TotalChangedLines)
	}

	if result.CoveredLines != 1 {
		t.Errorf("expected 1 covered line, got %d", result.CoveredLines)
	}

	if result.UncoveredLines != 1 {
		t.Errorf("expected 1 uncovered line, got %d", result.UncoveredLines)
	}

	expectedCoverage := 50.0
	if result.CoveragePercentage != expectedCoverage {
		t.Errorf("expected %.1f%% coverage, got %.1f%%", expectedCoverage, result.CoveragePercentage)
	}

	// Check file results
	fileResult, ok := result.FileResults["file.go"]
	if !ok {
		t.Fatal("expected file result for file.go")
	}

	if len(fileResult.UncoveredLineNumbers) != 1 {
		t.Errorf("expected 1 uncovered line number, got %d", len(fileResult.UncoveredLineNumbers))
	}

	if fileResult.UncoveredLineNumbers[0] != 7 {
		t.Errorf("expected uncovered line 7, got %d", fileResult.UncoveredLineNumbers[0])
	}

	if len(fileResult.CoveredLineNumbers) != 1 {
		t.Errorf("expected 1 covered line number, got %d", len(fileResult.CoveredLineNumbers))
	}

	if fileResult.CoveredLineNumbers[0] != 6 {
		t.Errorf("expected covered line 6, got %d", fileResult.CoveredLineNumbers[0])
	}

	// Verify new file detection (file.go should be detected as modified, not new)
	if fileResult.IsNewFile {
		t.Error("expected file.go to be detected as modified file, not new file")
	}

	// Verify metrics are initialized
	if result.NewFileMetrics == nil {
		t.Error("expected NewFileMetrics to be initialized")
	}
	if result.ModifiedFileMetrics == nil {
		t.Error("expected ModifiedFileMetrics to be initialized")
	}
}

func TestAnalyze_NoCoverage(t *testing.T) {
	diffOutput := `diff --git a/file.go b/file.go
index 123..456 100644
--- a/file.go
+++ b/file.go
@@ -5,1 +5,2 @@
+	new line
`

	diffResult, err := hunk.ParseGitDiff(diffOutput)
	if err != nil {
		t.Fatalf("failed to parse diff: %v", err)
	}

	// Empty coverage report
	coverageReport := &coverage.Report{
		FileCoverage: make(map[string]*coverage.CoverageData),
	}

	result, err := Analyze(diffResult, coverageReport)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.UncoveredLines != 1 {
		t.Errorf("expected 1 uncovered line, got %d", result.UncoveredLines)
	}

	if result.CoveragePercentage != 0.0 {
		t.Errorf("expected 0%% coverage, got %.1f%%", result.CoveragePercentage)
	}
}

func TestAnalysisResult_MeetsThreshold(t *testing.T) {
	result := &AnalysisResult{
		TotalChangedLines:  10,
		CoveredLines:       8,
		UncoveredLines:     2,
		CoveragePercentage: 80.0,
	}

	if !result.MeetsThreshold(80.0) {
		t.Error("expected to meet 80% threshold")
	}

	if !result.MeetsThreshold(75.0) {
		t.Error("expected to meet 75% threshold")
	}

	if result.MeetsThreshold(85.0) {
		t.Error("expected not to meet 85% threshold")
	}
}

func TestAnalysisResult_HasUncoveredLines(t *testing.T) {
	result1 := &AnalysisResult{
		UncoveredLines: 5,
	}
	if !result1.HasUncoveredLines() {
		t.Error("expected to have uncovered lines")
	}

	result2 := &AnalysisResult{
		UncoveredLines: 0,
	}
	if result2.HasUncoveredLines() {
		t.Error("expected not to have uncovered lines")
	}
}

func TestAnalyze_NewFile(t *testing.T) {
	// Create test diff for a new file
	diffOutput := `diff --git a/newfile.go b/newfile.go
new file mode 100644
index 0000000..abcdefg
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,3 @@
+package main
+
+func main() {}
`

	// Create test LCOV file with coverage for the new file
	lcovContent := `TN:
SF:newfile.go
DA:1,1
DA:3,1
end_of_record
`

	diffResult, err := hunk.ParseGitDiff(diffOutput)
	if err != nil {
		t.Fatalf("failed to parse diff: %v", err)
	}

	tmpfile, err := os.CreateTemp("", "test-*.info")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(lcovContent)); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}
	tmpfile.Close()

	coverageReport, err := coverage.ParseLCOV(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to parse coverage: %v", err)
	}

	result, err := Analyze(diffResult, coverageReport)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify new file is detected
	fileResult, ok := result.FileResults["newfile.go"]
	if !ok {
		t.Fatal("expected file result for newfile.go")
	}

	if !fileResult.IsNewFile {
		t.Error("expected newfile.go to be detected as new file")
	}

	// Verify new file metrics
	if result.NewFileMetrics.FileCount != 1 {
		t.Errorf("expected 1 new file, got %d", result.NewFileMetrics.FileCount)
	}

	if result.ModifiedFileMetrics.FileCount != 0 {
		t.Errorf("expected 0 modified files, got %d", result.ModifiedFileMetrics.FileCount)
	}
}

func TestAnalyzeWithBaseline(t *testing.T) {
	// Create test diff for a modified file
	diffOutput := `diff --git a/file.go b/file.go
index 123..456 100644
--- a/file.go
+++ b/file.go
@@ -5,3 +5,5 @@ func main() {
 	fmt.Println("hello")
+	fmt.Println("new line 1")
+	fmt.Println("new line 2")
 	fmt.Println("world")
`

	// Current coverage - line 6 covered, line 7 not covered
	currentCoverage := `TN:
SF:file.go
DA:6,5
DA:7,0
DA:8,3
end_of_record
`

	// Baseline coverage - line 6 was covered before, line 7 didn't exist
	baselineCoverage := `TN:
SF:file.go
DA:6,3
DA:8,2
end_of_record
`

	diffResult, err := hunk.ParseGitDiff(diffOutput)
	if err != nil {
		t.Fatalf("failed to parse diff: %v", err)
	}

	// Parse current coverage
	currentFile, err := os.CreateTemp("", "current-*.info")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(currentFile.Name())
	currentFile.Write([]byte(currentCoverage))
	currentFile.Close()

	currentReport, err := coverage.ParseLCOV(currentFile.Name())
	if err != nil {
		t.Fatalf("failed to parse current coverage: %v", err)
	}

	// Parse baseline coverage
	baselineFile, err := os.CreateTemp("", "baseline-*.info")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(baselineFile.Name())
	baselineFile.Write([]byte(baselineCoverage))
	baselineFile.Close()

	baselineReport, err := coverage.ParseLCOV(baselineFile.Name())
	if err != nil {
		t.Fatalf("failed to parse baseline coverage: %v", err)
	}

	result, err := AnalyzeWithBaseline(diffResult, currentReport, baselineReport)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fileResult, ok := result.FileResults["file.go"]
	if !ok {
		t.Fatal("expected file result for file.go")
	}

	// Verify baseline coverage is tracked
	// Line 6 existed in baseline and was covered (100% baseline for that line)
	// Line 7 is new, so baseline coverage should be 0% for that line
	// Overall baseline for changed lines: 1 out of 2 lines = 50%
	if fileResult.BaselineCoveragePercentage != 50.0 {
		t.Errorf("expected baseline coverage 50%%, got %.1f%%", fileResult.BaselineCoveragePercentage)
	}

	// Current coverage: 1 out of 2 lines = 50%
	if fileResult.CoveragePercentage != 50.0 {
		t.Errorf("expected current coverage 50%%, got %.1f%%", fileResult.CoveragePercentage)
	}
}
