# Difftron

**Difftron** is a language-agnostic, AI-powered "Quality Gate" CLI. It ensures that new code changes are adequately tested by correlating `git diff` hunks with standard coverage reports (LCOV, Cobertura, etc.).

Unlike traditional coverage tools that report on the entire project, **Difftron** zooms in on the **deltas**, ensuring that every new line of code is held to a high standard of quality before it hits production.

## The Problem

In large repositories, a developer can introduce 100 lines of untested logic, and the overall project coverage might only drop by 0.01%. This makes it easy for "testing debt" to accumulate unnoticed.

## The Solution

Difftron parses the `git diff` to identify the specific line numbers modified in a PR. It then "intersects" this data with your coverage reports to identify exactly which new lines are untested. Finally, it uses Gemini AI to suggest the missing test cases.

---

## Architecture

| Component | Responsibility |
| --- | --- |
| **Hunk Engine** | Parses `git diff` output to map "hunks" to absolute line numbers in the new file version. |
| **Coverage Engine** | Ingests LCOV/Cobertura files to build an in-memory map of `File -> Line -> HitCount`. |
| **Risk Engine** | Cross-references coverage with **Git Churn** (frequency of file changes) to flag high-risk gaps. |
| **Gemini Integration** | Provides "Agentic" test generation by sending uncovered hunks to the AI for boilerplate creation. |

---

## Implementation Highlights (Go)

### 1. The Hunk-to-Line Mapper

This core logic translates relative diff offsets into absolute line numbers that match your coverage report.

```go
// ParseGitDiff identifies absolute line numbers of added/modified code
func ParseGitDiff(diffOutput string) map[string]map[int]bool {
	changes := make(map[string]map[int]bool)
	scanner := bufio.NewScanner(strings.NewReader(diffOutput))
	var currentFile string
	var currentLine int

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			changes[currentFile] = make(map[int]bool)
			continue
		}

		// Header Example: @@ -10,5 +15,7 @@
		if strings.HasPrefix(line, "@@") {
			parts := strings.Split(line, " ")
			newFilePart := strings.Split(parts[2], ",") // e.g., +15
			startLine, _ := strconv.Atoi(newFilePart[0][1:])
			currentLine = startLine
			continue
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			changes[currentFile][currentLine] = true
			currentLine++
		} else if !strings.HasPrefix(line, "-") {
			currentLine++ // Context line
		}
	}
	return changes
}

```

### 2. The LCOV Coverage Parser

A high-performance scanner to ingest standardized coverage data.

```go
// ParseLCOV reads .info files into a File -> Line -> Hits map
func ParseLCOV(filePath string) (map[string]map[int]int, error) {
	file, _ := os.Open(filePath)
	defer file.Close()

	reports := make(map[string]map[int]int)
	var currentFile string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "SF:") {
			currentFile = line[3:]
			reports[currentFile] = make(map[int]int)
		} else if strings.HasPrefix(line, "DA:") {
			data := strings.Split(line[3:], ",")
			lineNum, _ := strconv.Atoi(data[0])
			hits, _ := strconv.Atoi(data[1])
			reports[currentFile][lineNum] = hits
		}
	}
	return reports, nil
}

```

---

## Key Features

* **Universal Language Support**: If your language exports LCOV or Cobertura (Go, TS, Java, Python, C++, etc.), Difftron can analyze it.
* **Risk Heatmaps**: Uses Git Churn data to identify "Hot Spot" files. Low coverage in a high-churn file triggers a **CRITICAL** alert.
* **AI Test Generation**: Don't just report the gap—fix it. Difftron generates `_test.go` or `.test.ts` snippets for missing paths.
* **CI/CD Native**: Designed to run as a GitHub Action or GitLab CI Job, posting rich Markdown reports directly to your PR/MR.
* **Dogfooding Ready**: Use difftron on itself! Built-in support for analyzing your own code changes.

---

## Development

### Prerequisites

- Go 1.21 or later
- Git (for testing git diff functionality)

### Setup

```bash
# Clone the repository
git clone <repo-url>
cd difftron

# Initialize Go module (if not already done)
go mod init github.com/swantron/difftron

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build the CLI
go build -o bin/difftron ./cmd/difftron

# Or use the task runner
go run scripts/task.go build
go run scripts/task.go test
```

### Task Runner

Difftron uses a Go-based task runner instead of a Makefile. This keeps the project 100% Go and avoids external dependencies.

```bash
# Build the binary
go run scripts/task.go build

# Run all tests
go run scripts/task.go test

# Run tests with coverage
go run scripts/task.go test-coverage

# Run integration tests with fixtures
go run scripts/task.go test-integration

# Dogfood: Analyze your own changes
go run scripts/task.go dogfood

# Install the binary
go run scripts/task.go install

# Format code
go run scripts/task.go fmt

# Clean build artifacts
go run scripts/task.go clean

# Run the CLI locally
go run scripts/task.go run analyze --coverage coverage.info
```

### Testing

```bash
# Run all unit tests
go run scripts/task.go test

# Run integration test with fixtures
go run scripts/task.go test-integration

# Run tests with coverage report
go run scripts/task.go test-coverage

# Or use go test directly
go test ./...
go test -v ./...
go test ./internal/hunk/...
```

See [TESTING.md](TESTING.md) for comprehensive testing guide.

---

## Dogfooding

Difftron can analyze itself! Use it to ensure your own code changes meet coverage standards.

```bash
# Analyze your last commit
go run scripts/task.go dogfood

# Analyze specific commit range
go run scripts/task.go dogfood HEAD~5 HEAD
```

See [DOGFOOD.md](DOGFOOD.md) for detailed dogfooding guide and CI/CD integration.

---

## Implementation Plan

### Phase 1: Foundation (v0.1) - CLI Core

**Goal**: Build the core CLI with git diff parsing and LCOV coverage analysis.

#### Tasks:
1. **Project Setup**
   - Initialize Go module (`go mod init`)
   - Set up project structure (`cmd/`, `internal/`, `pkg/`)
   - Add CLI framework (Cobra)
   - Create Go-based task runner for common tasks

2. **Hunk Engine** (`internal/hunk/`)
   - Implement `ParseGitDiff()` to extract changed lines
   - Handle unified diff format parsing
   - Support both staged and unstaged diffs
   - Map relative diff positions to absolute line numbers
   - Unit tests with sample diff outputs

3. **Coverage Engine** (`internal/coverage/`)
   - Implement `ParseLCOV()` for `.info` files
   - Build `File -> Line -> HitCount` data structure
   - Handle file path normalization (relative vs absolute)
   - Error handling for malformed LCOV files
   - Unit tests with sample LCOV data

4. **Core Analysis** (`internal/analyzer/`)
   - Intersect diff hunks with coverage data
   - Calculate coverage percentage for changed lines
   - Identify uncovered lines in diffs
   - Generate coverage report summary

5. **CLI Interface** (`cmd/difftron/`)
   - `difftron analyze` command
   - Flags: `--diff`, `--coverage`, `--threshold`, `--output`
   - JSON and human-readable output formats
   - Exit codes: 0 (pass), 1 (fail), 2 (error)

**Deliverables**: Working CLI that can analyze a git diff against LCOV coverage.

---

### Phase 2: Risk Scoring (v0.2)

**Goal**: Add git churn analysis and risk-based prioritization.

#### Tasks:
1. **Git Churn Analysis** (`internal/churn/`)
   - Calculate file change frequency
   - Analyze commit history for hot spots
   - Weight recent changes more heavily
   - Cache churn data for performance

2. **Risk Engine** (`internal/risk/`)
   - Combine coverage gaps with churn scores
   - Assign risk levels: LOW, MEDIUM, HIGH, CRITICAL
   - Prioritize uncovered lines in high-churn files
   - Generate risk heatmap report

3. **Enhanced Reporting**
   - Color-coded output (red/yellow/green)
   - Risk-based sorting of issues
   - File-level and line-level risk scores

**Deliverables**: Risk-aware coverage analysis with prioritized alerts.

---

### Phase 3: CI/CD Integration (v0.3)

**Goal**: Integrate with GitHub Actions and GitLab CI.

#### Tasks:
1. **GitHub Action** (`.github/workflows/`)
   - Action YAML template
   - Auto-detect coverage files
   - Post PR comments with coverage report
   - Support for pull_request events

2. **GitLab CI** (`.gitlab-ci.yml`)
   - CI job template
   - Merge request comment integration
   - Artifact generation

3. **Comment Formatting**
   - Markdown report generation
   - Collapsible sections for large diffs
   - Links to uncovered lines
   - Summary statistics

**Deliverables**: Ready-to-use CI/CD templates for GitHub and GitLab.

---

### Phase 4: AI Test Generation (v0.4)

**Goal**: Generate test code suggestions using Gemini AI.

#### Tasks:
1. **Gemini Integration** (`internal/ai/`)
   - Google AI SDK integration
   - Prompt engineering for test generation
   - Context-aware code suggestions
   - Support for multiple languages (Go, TypeScript, Python, etc.)

2. **Test Generation Engine**
   - Extract uncovered code snippets
   - Generate language-specific test templates
   - Include imports and setup code
   - Format output as code blocks

3. **CLI Enhancement**
   - `difftron generate` command
   - `--ai-provider` flag (default: gemini)
   - `--language` flag for test generation
   - `--output-file` for saving generated tests

**Deliverables**: AI-powered test generation for uncovered code paths.

---

## Project Structure

```
difftron/
├── cmd/
│   └── difftron/
│       └── main.go              # CLI entry point
├── internal/
│   ├── hunk/
│   │   ├── parser.go            # Git diff parsing
│   │   └── parser_test.go
│   ├── coverage/
│   │   ├── lcov.go              # LCOV parser
│   │   ├── gocov.go             # Go coverage parser
│   │   ├── cobertura.go         # Cobertura parser (v0.2)
│   │   └── coverage_test.go
│   ├── analyzer/
│   │   ├── analyzer.go          # Core analysis logic
│   │   └── analyzer_test.go
│   ├── churn/
│   │   ├── churn.go             # Git churn calculation (v0.2)
│   │   └── churn_test.go
│   ├── risk/
│   │   ├── scorer.go            # Risk scoring (v0.2)
│   │   └── scorer_test.go
│   └── ai/
│       ├── gemini.go            # Gemini integration (v0.4)
│       └── generator.go         # Test generation (v0.4)
├── pkg/
│   └── report/
│       ├── formatter.go         # Report formatting
│       └── markdown.go          # Markdown generation
├── scripts/
│   ├── task.go                  # Go-based task runner
│   ├── dogfood.sh               # Dogfooding script
│   └── generate-coverage.sh     # Coverage generation helper
├── testdata/
│   └── fixtures/                # Test fixtures
├── .github/
│   └── workflows/
│       └── difftron.yml         # GitHub Action (v0.3)
├── .gitlab-ci.yml               # GitLab CI template (v0.3)
├── go.mod
├── go.sum
├── README.md
├── TESTING.md                   # Testing guide
├── DOGFOOD.md                   # Dogfooding guide
└── LICENSE
```

---

## Quick Start (After v0.1)

```bash
# Install
go install github.com/swantron/difftron/cmd/difftron@latest

# Analyze current diff against coverage
difftron analyze --coverage coverage.info

# Analyze specific diff
git diff main...feature-branch | difftron analyze --coverage coverage.info

# Set coverage threshold
difftron analyze --coverage coverage.info --threshold 80

# Generate JSON report
difftron analyze --coverage coverage.info --output json > report.json
```

---

## Roadmap Summary

* [x] **v0.1**: CLI Core (Diff + LCOV parsing). **COMPLETE**
* [ ] **v0.2**: Risk Scoring (Git Churn + Complexity).
* [ ] **v0.3**: GitHub Action/GitLab CI Commenter.
* [ ] **v0.4**: Gemini-powered Test Generation.

---

## Current Status (v0.1)

**Phase 1 is complete!** The core CLI functionality is implemented and tested:

### Implemented Features:
- Git diff parsing (Hunk Engine)
- LCOV coverage file parsing (Coverage Engine)
- Go coverage format support (for dogfooding)
- Core analysis engine (intersects diffs with coverage)
- CLI interface with `analyze` command
- Text and JSON output formats
- Coverage threshold checking
- Comprehensive test coverage
- Dogfooding support (analyze your own code)

### Usage Example:

```bash
# Analyze current working directory changes
difftron analyze --coverage coverage.info

# Analyze with custom threshold
difftron analyze --coverage coverage.info --threshold 90

# Analyze specific git diff range
difftron analyze --coverage coverage.info --base main --head feature-branch

# JSON output
difftron analyze --coverage coverage.info --output json

# Use a pre-generated diff file
git diff main...feature > diff.patch
difftron analyze --coverage coverage.info --diff diff.patch

# Test with included fixtures
difftron analyze --coverage testdata/fixtures/tronswan-coverage.info --diff testdata/fixtures/sample.diff

# Dogfood: Analyze your own changes
go run scripts/task.go dogfood
```

### Next Steps:
- Implement Risk Engine (git churn analysis) for v0.2
- Add Cobertura XML parser support
- Create CI/CD integration templates
- Add Gemini AI integration for test generation

---
