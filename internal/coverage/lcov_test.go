package coverage

import (
	"os"
	"testing"
)

func TestParseLCOV(t *testing.T) {
	// Create a temporary LCOV file
	lcovContent := `TN:
SF:file1.go
DA:10,5
DA:11,3
DA:12,0
DA:13,1
end_of_record
TN:
SF:file2.go
DA:5,10
DA:6,0
DA:7,2
end_of_record
`

	tmpfile, err := os.CreateTemp("", "test-*.info")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(lcovContent)); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}
	tmpfile.Close()

	// Parse the LCOV file
	report, err := ParseLCOV(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file1.go coverage
	file1Coverage := report.GetCoverageForFile("file1.go")
	if file1Coverage == nil {
		t.Fatal("expected coverage data for file1.go")
	}

	if file1Coverage.TotalLines != 4 {
		t.Errorf("expected 4 total lines for file1.go, got %d", file1Coverage.TotalLines)
	}

	if file1Coverage.CoveredLines != 3 {
		t.Errorf("expected 3 covered lines for file1.go, got %d", file1Coverage.CoveredLines)
	}

	// Check specific line coverage
	if report.GetCoverageForLine("file1.go", 10) != 5 {
		t.Errorf("expected line 10 to have 5 hits, got %d", report.GetCoverageForLine("file1.go", 10))
	}

	if report.GetCoverageForLine("file1.go", 12) != 0 {
		t.Errorf("expected line 12 to have 0 hits, got %d", report.GetCoverageForLine("file1.go", 12))
	}

	if report.IsLineCovered("file1.go", 10) != true {
		t.Error("expected line 10 to be covered")
	}

	if report.IsLineCovered("file1.go", 12) != false {
		t.Error("expected line 12 to not be covered")
	}

	// Check file2.go coverage
	file2Coverage := report.GetCoverageForFile("file2.go")
	if file2Coverage == nil {
		t.Fatal("expected coverage data for file2.go")
	}

	if file2Coverage.TotalLines != 3 {
		t.Errorf("expected 3 total lines for file2.go, got %d", file2Coverage.TotalLines)
	}

	// Check non-existent file
	if report.GetCoverageForFile("nonexistent.go") != nil {
		t.Error("expected nil for non-existent file")
	}

	if report.GetCoverageForLine("nonexistent.go", 1) != 0 {
		t.Error("expected 0 hits for non-existent file")
	}
}

func TestParseLCOV_EmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-empty-*.info")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	report, err := ParseLCOV(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.FileCoverage) != 0 {
		t.Errorf("expected empty report, got %d files", len(report.FileCoverage))
	}
}

func TestParseLCOV_InvalidFile(t *testing.T) {
	_, err := ParseLCOV("/nonexistent/file.info")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.go", "file.go"},
		{"./file.go", "file.go"},
		{"/file.go", "file.go"},
		{"./path/to/file.go", "path/to/file.go"},
		{"/absolute/path/file.go", "absolute/path/file.go"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
