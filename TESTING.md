# Testing Guide

This guide covers how to test Difftron at different levels.

## Quick Start

```bash
# Run all unit tests
go run scripts/task.go test

# Run integration test with fixtures
go run scripts/task.go test-integration

# Run tests with coverage report
go run scripts/task.go test-coverage
```

---

## Unit Tests

Unit tests verify individual components in isolation.

### Run All Unit Tests

```bash
# Using the task runner
go run scripts/task.go test

# Or directly with go test
go test ./...

# Verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/hunk/...
go test ./internal/coverage/...
go test ./internal/analyzer/...
```

### Test Coverage

Generate a coverage report:

```bash
# Generate coverage report and HTML
go run scripts/task.go test-coverage

# This creates:
# - coverage.out (raw coverage data)
# - coverage.html (visual HTML report)

# View HTML report (open in browser)
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Individual Test Files

Each package has its own test file:
- `internal/hunk/parser_test.go` - Tests git diff parsing
- `internal/coverage/lcov_test.go` - Tests LCOV parsing
- `internal/analyzer/analyzer_test.go` - Tests analysis logic

---

## Integration Tests

Integration tests verify the CLI works end-to-end with real data.

### Using Test Fixtures

The project includes test fixtures in `testdata/fixtures/`:

```bash
# Run the built-in integration test
go run scripts/task.go test-integration

# Or manually test with fixtures
./bin/difftron analyze \
  --coverage testdata/fixtures/tronswan-coverage.info \
  --diff testdata/fixtures/sample.diff \
  --threshold 0
```

### Test Fixtures

- `testdata/fixtures/tronswan-coverage.info` - Real LCOV coverage from tronswan repo
- `testdata/fixtures/sample.diff` - Sample git diff for testing

---

## Manual CLI Testing

### Build the Binary

```bash
go run scripts/task.go build
# Binary will be at: bin/difftron
```

### Test Scenarios

#### 1. Test with Current Git Changes

```bash
# Analyze your current working directory changes
./bin/difftron analyze --coverage coverage.info

# Or if you have a coverage file
./bin/difftron analyze --coverage testdata/fixtures/tronswan-coverage.info
```

#### 2. Test with a Specific Diff File

```bash
# Create a test diff
git diff HEAD~1 HEAD > test.diff

# Analyze it
./bin/difftron analyze \
  --coverage testdata/fixtures/tronswan-coverage.info \
  --diff test.diff
```

#### 3. Test with Different Thresholds

```bash
# Should pass (0% threshold)
./bin/difftron analyze \
  --coverage testdata/fixtures/tronswan-coverage.info \
  --diff testdata/fixtures/sample.diff \
  --threshold 0

# Should fail (100% threshold)
./bin/difftron analyze \
  --coverage testdata/fixtures/tronswan-coverage.info \
  --diff testdata/fixtures/sample.diff \
  --threshold 100
```

#### 4. Test JSON Output

```bash
./bin/difftron analyze \
  --coverage testdata/fixtures/tronswan-coverage.info \
  --diff testdata/fixtures/sample.diff \
  --output json
```

#### 5. Test with Git Diff Range

```bash
# Compare two branches
./bin/difftron analyze \
  --coverage testdata/fixtures/tronswan-coverage.info \
  --base main \
  --head feature-branch
```

---

## Testing with Real Projects

### Generate Coverage from Another Project

#### For Go Projects:

```bash
cd /path/to/go-project
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out > coverage.info

# Copy to difftron
cp coverage.info /path/to/difftron/testdata/fixtures/
```

#### For TypeScript/JavaScript Projects (using Vitest):

```bash
cd /path/to/typescript-project
yarn test:coverage
# This generates coverage/lcov.info

# Copy to difftron
cp coverage/lcov.info /path/to/difftron/testdata/fixtures/
```

### Test Against Real Diff

```bash
# In your project
cd /path/to/your-project
git diff main...feature-branch > /tmp/real-diff.patch

# In difftron
cd /path/to/difftron
./bin/difftron analyze \
  --coverage testdata/fixtures/your-coverage.info \
  --diff /tmp/real-diff.patch
```

---

## Test Scenarios Checklist

### Basic Functionality

- [ ] Parse simple git diff
- [ ] Parse complex git diff with multiple files
- [ ] Parse LCOV coverage file
- [ ] Match diff lines with coverage data
- [ ] Calculate coverage percentage correctly
- [ ] Identify uncovered lines

### Edge Cases

- [ ] Empty diff (no changes)
- [ ] Empty coverage file
- [ ] File in diff but not in coverage
- [ ] File in coverage but not in diff
- [ ] Path normalization (relative vs absolute)
- [ ] Multiple hunks in same file
- [ ] Deleted files in diff

### CLI Features

- [ ] Text output format
- [ ] JSON output format
- [ ] Threshold checking (pass/fail)
- [ ] Exit codes (0=pass, 1=fail, 2=error)
- [ ] Help text displays correctly
- [ ] Error handling for missing files
- [ ] Error handling for invalid arguments

### Integration

- [ ] Works with real LCOV files
- [ ] Works with real git diffs
- [ ] Handles large coverage files
- [ ] Handles large diffs

---

## Continuous Testing

### Run Tests Before Committing

```bash
# Format code
go run scripts/task.go fmt

# Run tests
go run scripts/task.go test

# Check coverage
go run scripts/task.go test-coverage
```

### CI/CD Testing

Add to your CI pipeline:

```yaml
# Example GitHub Actions
- name: Run tests
  run: go test ./...

- name: Run integration test
  run: go run scripts/task.go test-integration
```

---

## Debugging Tests

### Run Specific Test

```bash
# Run a specific test function
go test -v ./internal/hunk -run TestParseGitDiff

# Run tests matching a pattern
go test -v ./internal/analyzer -run TestAnalyze
```

### Verbose Output

```bash
# See detailed test output
go test -v ./...

# See coverage for each function
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Debug with Delve

```bash
# Install delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a test
dlv test ./internal/hunk -test.run TestParseGitDiff
```

---

## Writing New Tests

### Test File Structure

```go
package hunk

import (
    "testing"
)

func TestNewFeature(t *testing.T) {
    // Arrange
    input := "test input"
    
    // Act
    result := ParseGitDiff(input)
    
    // Assert
    if result == nil {
        t.Error("Expected non-nil result")
    }
}
```

### Test Table Pattern

```go
func TestParseGitDiff(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        wantError   bool
        wantFiles   []string
    }{
        {
            name:      "simple case",
            input:     "diff --git a/file.go...",
            wantError: false,
            wantFiles: []string{"file.go"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

---

## Troubleshooting

### Tests Fail Unexpectedly

1. Check if fixtures exist:
   ```bash
   ls -la testdata/fixtures/
   ```

2. Regenerate coverage if needed:
   ```bash
   cd /path/to/tronswan
   yarn test:coverage
   cp coverage/lcov.info /path/to/difftron/testdata/fixtures/tronswan-coverage.info
   ```

3. Check Go version:
   ```bash
   go version  # Should be 1.21+
   ```

### Coverage Report Not Generated

```bash
# Ensure coverage.out exists
ls -la coverage.out

# Regenerate if needed
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Binary Not Found

```bash
# Build first
go run scripts/task.go build

# Check if binary exists
ls -la bin/difftron
```

---

## Next Steps

- Add more test fixtures for different scenarios
- Add performance benchmarks
- Add fuzz testing for parsers
- Add end-to-end tests with real git repos
