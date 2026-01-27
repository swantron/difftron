package health

import (
	"testing"

	"github.com/swantron/difftron/internal/coverage"
	"github.com/swantron/difftron/internal/hunk"
)

func TestAggregateCoverage(t *testing.T) {
	t.Run("empty reports", func(t *testing.T) {
		result, err := AggregateCoverage([]*TestCoverageReport{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
		if result.FileCoverage == nil {
			t.Error("expected FileCoverage map")
		}
	})

	t.Run("single report", func(t *testing.T) {
		report := &TestCoverageReport{
			TestType: TestTypeUnit,
			CoverageReport: &coverage.Report{
				FileCoverage: map[string]*coverage.CoverageData{
					"file.go": {
						TotalLines:   10,
						CoveredLines: 8,
						LineHits: map[int]int{
							1: 5, 2: 3, 3: 0, 4: 2,
						},
					},
				},
			},
		}

		result, err := AggregateCoverage([]*TestCoverageReport{report})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fileCov := result.FileCoverage["file.go"]
		if fileCov == nil {
			t.Fatal("expected file coverage")
		}
		if fileCov.TotalLines != 10 {
			t.Errorf("expected 10 total lines, got %d", fileCov.TotalLines)
		}
		if fileCov.CoveredLines != 3 {
			t.Errorf("expected 3 covered lines, got %d", fileCov.CoveredLines)
		}
	})

	t.Run("multiple reports merge", func(t *testing.T) {
		report1 := &TestCoverageReport{
			TestType: TestTypeUnit,
			CoverageReport: &coverage.Report{
				FileCoverage: map[string]*coverage.CoverageData{
					"file.go": {
						TotalLines:   10,
						CoveredLines: 5,
						LineHits: map[int]int{
							1: 5, 2: 3, 3: 0,
						},
					},
				},
			},
		}

		report2 := &TestCoverageReport{
			TestType: TestTypeAPI,
			CoverageReport: &coverage.Report{
				FileCoverage: map[string]*coverage.CoverageData{
					"file.go": {
						TotalLines:   10,
						CoveredLines: 4,
						LineHits: map[int]int{
							3: 2, 4: 1, 5: 0,
						},
					},
				},
			},
		}

		result, err := AggregateCoverage([]*TestCoverageReport{report1, report2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fileCov := result.FileCoverage["file.go"]
		if fileCov == nil {
			t.Fatal("expected file coverage")
		}
		// Line 3 should be covered (by report2) even though report1 has it as 0
		if fileCov.LineHits[3] != 2 {
			t.Errorf("expected line 3 to have 2 hits, got %d", fileCov.LineHits[3])
		}
		// Should take maximum hit count
		if fileCov.LineHits[1] != 5 {
			t.Errorf("expected line 1 to have 5 hits, got %d", fileCov.LineHits[1])
		}
	})

	t.Run("nil coverage report skipped", func(t *testing.T) {
		report := &TestCoverageReport{
			TestType:       TestTypeUnit,
			CoverageReport: nil,
		}

		result, err := AggregateCoverage([]*TestCoverageReport{report})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
	})
}

func TestAnalyzeHealth(t *testing.T) {
	t.Run("nil diff result", func(t *testing.T) {
		reports := []*TestCoverageReport{
			{
				TestType: TestTypeUnit,
				CoverageReport: &coverage.Report{
					FileCoverage: make(map[string]*coverage.CoverageData),
				},
			},
		}

		_, err := AnalyzeHealth(nil, reports, nil, 80.0)
		if err == nil {
			t.Error("expected error for nil diff result")
		}
	})

	t.Run("empty test reports", func(t *testing.T) {
		diffResult := &hunk.ParseResult{
			ChangedLines: make(map[string]map[int]bool),
			AddedLines:   make(map[string]map[int]bool),
			RemovedLines: make(map[string]map[int]bool),
			NewFiles:     make(map[string]bool),
		}

		_, err := AnalyzeHealth(diffResult, []*TestCoverageReport{}, nil, 80.0)
		if err == nil {
			t.Error("expected error for empty test reports")
		}
	})

	t.Run("valid analysis", func(t *testing.T) {
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

		diffResult, err := hunk.ParseGitDiff(diffOutput)
		if err != nil {
			t.Fatalf("failed to parse diff: %v", err)
		}

		testReports := []*TestCoverageReport{
			{
				TestType: TestTypeUnit,
				CoverageReport: &coverage.Report{
					FileCoverage: map[string]*coverage.CoverageData{
						"file.go": {
							TotalLines:   10,
							CoveredLines: 8,
							LineHits: map[int]int{
								6: 5, 7: 3, 8: 2,
							},
						},
					},
				},
			},
		}

		report, err := AnalyzeHealth(diffResult, testReports, nil, 80.0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report == nil {
			t.Fatal("expected non-nil report")
		}
		if report.ChangedFiles != 1 {
			t.Errorf("expected 1 changed file, got %d", report.ChangedFiles)
		}
	})
}

func TestHealthReport_calculateOverallMetrics(t *testing.T) {
	report := &HealthReport{
		FileHealth: make(map[string]*FileHealth),
	}

	coverageReport := &coverage.Report{
		FileCoverage: map[string]*coverage.CoverageData{
			"file1.go": {
				TotalLines:   10,
				CoveredLines: 8,
			},
			"file2.go": {
				TotalLines:   20,
				CoveredLines: 15,
			},
		},
	}

	report.calculateOverallMetrics(coverageReport)

	if report.TotalFiles != 2 {
		t.Errorf("expected 2 files, got %d", report.TotalFiles)
	}
	if report.TotalLines != 30 {
		t.Errorf("expected 30 total lines, got %d", report.TotalLines)
	}
	if report.TotalCoveredLines != 23 {
		t.Errorf("expected 23 covered lines, got %d", report.TotalCoveredLines)
	}
	expectedCoverage := float64(23) / float64(30) * 100
	if report.OverallCoverage != expectedCoverage {
		t.Errorf("expected coverage %.2f, got %.2f", expectedCoverage, report.OverallCoverage)
	}
}
