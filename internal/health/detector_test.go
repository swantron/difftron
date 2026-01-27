package health

import (
	"testing"

	"github.com/swantron/difftron/internal/coverage"
)

func TestDetectTestType(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected TestType
	}{
		{"unit test file", "unit_test.go", TestTypeUnit},
		{"unit in path", "unit/coverage.out", TestTypeUnit},
		{"go test command", "go test ./...", TestTypeUnit},
		{"pytest command", "pytest tests/", TestTypeUnit},
		{"jest command", "jest test", TestTypeUnit},
		{"api test file", "api_test.go", TestTypeUnit}, // _test.go matches unit first
		{"api in path", "api/coverage.xml", TestTypeAPI},
		{"integration in path", "integration/coverage.xml", TestTypeAPI},
		{"postman", "postman collection", TestTypeAPI},
		{"newman", "newman run", TestTypeAPI},
		{"functional test file", "functional_test.go", TestTypeUnit}, // _test.go matches unit first
		{"e2e test file", "e2e_test.go", TestTypeUnit}, // _test.go matches unit first
		{"end-to-end", "end-to-end tests", TestTypeFunctional},
		{"cypress", "cypress run", TestTypeFunctional},
		{"playwright", "playwright test", TestTypeFunctional},
		{"go coverage file", "coverage.out", TestTypeUnit},
		{"lcov file", "coverage.info", TestTypeUnit},
		{"unknown defaults to unit", "unknown_source", TestTypeUnit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectTestType(tt.source)
			if result != tt.expected {
				t.Errorf("DetectTestType(%q) = %v, want %v", tt.source, result, tt.expected)
			}
		})
	}
}

func TestParseTestCoverageReports(t *testing.T) {
	t.Run("matching lengths", func(t *testing.T) {
		coverageFiles := []string{"unit.out", "api.xml"}
		testTypes := []TestType{TestTypeUnit, TestTypeAPI}
		coverageReports := []*coverage.Report{
			{FileCoverage: make(map[string]*coverage.CoverageData)},
			{FileCoverage: make(map[string]*coverage.CoverageData)},
		}

		reports, err := ParseTestCoverageReports(coverageFiles, testTypes, coverageReports)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(reports) != 2 {
			t.Errorf("expected 2 reports, got %d", len(reports))
		}
		if reports[0].TestType != TestTypeUnit {
			t.Errorf("expected TestTypeUnit, got %v", reports[0].TestType)
		}
		if reports[1].TestType != TestTypeAPI {
			t.Errorf("expected TestTypeAPI, got %v", reports[1].TestType)
		}
	})

	t.Run("mismatched lengths", func(t *testing.T) {
		coverageFiles := []string{"unit.out"}
		testTypes := []TestType{TestTypeUnit}
		coverageReports := []*coverage.Report{
			{FileCoverage: make(map[string]*coverage.CoverageData)},
			{FileCoverage: make(map[string]*coverage.CoverageData)},
		}

		_, err := ParseTestCoverageReports(coverageFiles, testTypes, coverageReports)
		if err == nil {
			t.Error("expected error for mismatched lengths")
		}
	})

	t.Run("auto-detect test type", func(t *testing.T) {
		coverageFiles := []string{"unit_test.go", "api/coverage.xml"}
		testTypes := []TestType{"", ""} // Empty to trigger auto-detection
		coverageReports := []*coverage.Report{
			{FileCoverage: make(map[string]*coverage.CoverageData)},
			{FileCoverage: make(map[string]*coverage.CoverageData)},
		}

		reports, err := ParseTestCoverageReports(coverageFiles, testTypes, coverageReports)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reports[0].TestType != TestTypeUnit {
			t.Errorf("expected auto-detected TestTypeUnit, got %v", reports[0].TestType)
		}
		if reports[1].TestType != TestTypeAPI {
			t.Errorf("expected auto-detected TestTypeAPI, got %v", reports[1].TestType)
		}
	})
}
