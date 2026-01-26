package health

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/swantron/difftron/internal/coverage"
)

// DetectTestType attempts to detect the test type from a coverage source
// This can be a file path, test command, or other identifier
func DetectTestType(source string) TestType {
	source = strings.ToLower(source)

	// Check file paths
	if strings.Contains(source, "unit") || strings.Contains(source, "_test.go") {
		return TestTypeUnit
	}
	if strings.Contains(source, "api") || strings.Contains(source, "integration") {
		return TestTypeAPI
	}
	if strings.Contains(source, "functional") || strings.Contains(source, "e2e") || strings.Contains(source, "end-to-end") {
		return TestTypeFunctional
	}

	// Check file extensions
	ext := filepath.Ext(source)
	switch ext {
	case ".out":
		// Go coverage files are typically unit tests
		return TestTypeUnit
	case ".info":
		// LCOV files could be any type, default to unit
		return TestTypeUnit
	}

	// Check for common test command patterns
	if strings.Contains(source, "go test") {
		return TestTypeUnit
	}
	if strings.Contains(source, "pytest") || strings.Contains(source, "jest") {
		return TestTypeUnit
	}
	if strings.Contains(source, "cypress") || strings.Contains(source, "playwright") {
		return TestTypeFunctional
	}
	if strings.Contains(source, "postman") || strings.Contains(source, "newman") {
		return TestTypeAPI
	}

	// Default to unit tests
	return TestTypeUnit
}

// ParseTestCoverageReports creates TestCoverageReport instances from coverage files
// It attempts to detect test types from file paths or names
// Note: coverageReport should be pre-parsed using coverage.ParseLCOV() or coverage.ParseGoCoverage()
// This function just wraps it with test type information
func ParseTestCoverageReports(coverageFiles []string, testTypes []TestType, coverageReports []*coverage.Report) ([]*TestCoverageReport, error) {
	if len(coverageFiles) != len(coverageReports) {
		return nil, fmt.Errorf("coverageFiles and coverageReports must have the same length")
	}

	reports := make([]*TestCoverageReport, 0, len(coverageFiles))

	for i, file := range coverageFiles {
		var testType TestType
		if i < len(testTypes) && testTypes[i] != "" {
			testType = testTypes[i]
		} else {
			testType = DetectTestType(file)
		}

		reports = append(reports, &TestCoverageReport{
			TestType:      testType,
			CoverageReport: coverageReports[i],
			Source:        file,
		})
	}

	return reports, nil
}
