# Holistic Testing Health Reporting

## Overview

Difftron now provides comprehensive testing health analysis that goes beyond simple coverage percentages. It aggregates coverage from multiple test types (unit, API, functional) and provides actionable insights for engineers and AI agents.

## Key Principles

1. **Understand Changes AND Entirety**: Analyzes both changed code and overall project health
2. **Prevent Coverage Drops**: Tracks baseline coverage to detect regressions
3. **Holistic View**: Aggregates coverage across all test types
4. **Actionable Insights**: Provides clear recommendations for improvement
5. **AI-Friendly Format**: Structured output for both humans and AI agents

## Test Types Supported

- **Unit Tests**: Fast, granular tests for individual functions/components
- **API Tests**: Integration tests for API endpoints and services
- **Functional/E2E Tests**: End-to-end tests for complete workflows

## Usage

### Basic Usage

```go
import (
    "github.com/swantron/difftron/internal/health"
    "github.com/swantron/difftron/internal/coverage"
    "github.com/swantron/difftron/internal/hunk"
)

// Parse git diff
diffResult, _ := hunk.ParseGitDiff(diffOutput)

// Parse multiple coverage reports
unitCoverage, _ := coverage.ParseLCOV("unit-coverage.info")
apiCoverage, _ := coverage.ParseLCOV("api-coverage.info")
functionalCoverage, _ := coverage.ParseLCOV("functional-coverage.info")

// Create test coverage reports
testReports := []*health.TestCoverageReport{
    {
        TestType: health.TestTypeUnit,
        CoverageReport: unitCoverage,
        Source: "unit-coverage.info",
    },
    {
        TestType: health.TestTypeAPI,
        CoverageReport: apiCoverage,
        Source: "api-coverage.info",
    },
    {
        TestType: health.TestTypeFunctional,
        CoverageReport: functionalCoverage,
        Source: "functional-coverage.info",
    },
}

// Optional: Parse baseline coverage
baselineReports := []*health.TestCoverageReport{
    {
        TestType: health.TestTypeUnit,
        CoverageReport: baselineUnitCoverage,
        Source: "baseline-unit-coverage.info",
    },
}

// Analyze health
healthReport, err := health.AnalyzeHealth(
    diffResult,
    testReports,
    baselineReports,
    80.0, // threshold
)
```

### Output Formats

#### JSON (for AI agents and programmatic access)

```go
jsonOutput, _ := healthReport.ToJSON()
fmt.Println(string(jsonOutput))
```

#### Markdown (for human-readable reports)

```go
markdownOutput := healthReport.ToMarkdown()
fmt.Println(markdownOutput)
```

#### Structured Text (for AI agents)

```go
textOutput := healthReport.ToStructuredText()
fmt.Println(textOutput)
```

## Health Report Structure

### Summary Metrics

- **Overall Coverage**: Project-wide coverage percentage
- **Changed Coverage**: Coverage for changed lines only
- **File Counts**: Total files, changed files, healthy/at-risk/regressing files
- **New vs Modified**: Separate metrics for new and modified files

### Test Type Breakdown

- **Unit Test Coverage**: Percentage covered by unit tests
- **API Test Coverage**: Percentage covered by API tests
- **Functional Test Coverage**: Percentage covered by functional tests

### File-Level Health

For each file:
- Overall coverage percentage
- Changed lines coverage percentage
- Baseline coverage (for modified files)
- Coverage delta (current - baseline)
- Test type breakdown (which test types cover changed lines)
- Status (healthy, at-risk, regressing)
- Uncovered line numbers

### Insights

Automatically generated insights include:
- Coverage threshold status
- Regression detection
- Test type gaps (e.g., only unit tests, no API tests)
- Missing coverage warnings

### Recommendations

Actionable recommendations with:
- Priority level (low, medium, high, critical)
- Category (add-tests, improve-coverage, fix-regression)
- Specific action to take
- Files needing attention
- Recommended test type

## Example Output

### Markdown Report

```markdown
# Testing Health Report

## Summary

**Overall Coverage:** 85.2%
**Changed Coverage:** 78.5%
**Total Files:** 42 | **Changed Files:** 3
**Healthy:** 2 | **At Risk:** 1 | **Regressing:** 0

## Test Type Coverage

| Test Type | Coverage |
|-----------|----------|
| Unit Tests | 82.3% |
| API Tests | 75.1% |
| Functional Tests | 68.9% |

## File Health

| File | Status | Changed Coverage | Baseline | Delta | Test Types |
|------|--------|-----------------|----------|-------|------------|
| `api/handlers.go` | ✅ Healthy | 95.0% | 90.0% | +5.0% | Unit, API |
| `internal/service.go` | ⚠️ At Risk | 65.0% | 70.0% | -5.0% | Unit |
| `pkg/utils.go` | ✅ Healthy | 88.0% | N/A | N/A | Unit, API |

## Insights

### ✅ Coverage threshold met

Overall change coverage is 78.5%, meeting the 80% threshold

### ⚠️ Unit tests only, consider integration tests

`internal/service.go` is covered by unit tests but not API/functional tests

## Recommendations

### ⚠️ Add tests for uncovered changes (high priority)

1 file(s) have changes below coverage threshold

**Action:** Add unit tests for uncovered changed lines

**Files needing attention:**
- `internal/service.go`
```

### JSON Output

```json
{
  "summary": {
    "overall_coverage": 85.2,
    "changed_coverage": 78.5,
    "total_files": 42,
    "changed_files": 3,
    "healthy_files": 2,
    "at_risk_files": 1,
    "regressing_files": 0,
    "new_files_count": 1,
    "modified_files_count": 2,
    "new_files_coverage": 88.0,
    "modified_files_coverage": 75.0
  },
  "test_types": {
    "unit_test_coverage": 82.3,
    "api_test_coverage": 75.1,
    "functional_test_coverage": 68.9
  },
  "changes": {
    "total_changed_lines": 120,
    "covered_lines": 94,
    "uncovered_lines": 26,
    "coverage_percentage": 78.5
  },
  "files": [
    {
      "file_path": "api/handlers.go",
      "is_new_file": false,
      "overall_coverage": 92.0,
      "changed_coverage": 95.0,
      "baseline_coverage": 90.0,
      "coverage_delta": 5.0,
      "changed_lines": 20,
      "covered_lines": 19,
      "uncovered_lines": 1,
      "unit_test_coverage": 95.0,
      "api_test_coverage": 90.0,
      "functional_test_coverage": 0.0,
      "status": "healthy"
    }
  ],
  "insights": [
    {
      "type": "success",
      "category": "coverage",
      "title": "Coverage threshold met",
      "description": "Overall change coverage is 78.5%, meeting the 80% threshold",
      "severity": "low"
    }
  ],
  "recommendations": [
    {
      "priority": "high",
      "category": "add-tests",
      "title": "Add tests for uncovered changes",
      "description": "1 file(s) have changes below coverage threshold",
      "action": "Add unit tests for uncovered changed lines",
      "files": ["internal/service.go"],
      "test_type": "unit"
    }
  ]
}
```

## Benefits

1. **Comprehensive View**: See coverage across all test types, not just unit tests
2. **Accurate Assessment**: Only evaluate changed code, with baseline comparison
3. **Actionable Insights**: Clear recommendations on what to fix and how
4. **AI-Friendly**: Structured formats that AI agents can parse and act on
5. **Regression Prevention**: Detect when coverage drops below baseline
6. **Test Gap Detection**: Identify when code is only covered by one test type

## Integration with CI/CD

### GitHub Actions Example

```yaml
- name: Run unit tests
  run: go test -coverprofile=unit-coverage.out ./...

- name: Run API tests
  run: pytest --cov=api --cov-report=lcov api-coverage.info

- name: Run functional tests
  run: npm run test:e2e -- --coverage

- name: Analyze health
  run: |
    ./bin/difftron health \
      --unit-coverage unit-coverage.out \
      --api-coverage api-coverage.info \
      --functional-coverage functional-coverage.info \
      --baseline-coverage baseline-coverage.json \
      --threshold 80 \
      --output json > health-report.json \
      --output markdown > health-report.md
```

## Future Enhancements

1. **Automatic Test Type Detection**: Infer test types from test file patterns
2. **Coverage Trend Tracking**: Track coverage over time per file
3. **Smart Recommendations**: AI-powered suggestions based on code patterns
4. **Integration with Test Frameworks**: Direct integration with popular test runners
5. **Coverage Overlap Analysis**: Identify redundant test coverage
6. **Performance Impact**: Track test execution time and optimize slow tests
