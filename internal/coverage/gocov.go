package coverage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ConvertGoCoverageToLCOV converts Go's coverage.out format to LCOV format
// This allows difftron to analyze its own Go code coverage
func ConvertGoCoverageToLCOV(coverageOutPath, outputPath string) error {
	// Read the coverage.out file to verify it exists
	_, err := os.ReadFile(coverageOutPath)
	if err != nil {
		return fmt.Errorf("failed to read coverage file: %w", err)
	}

	// Use go tool cover to convert to LCOV format
	// First, get the coverage data in a parseable format
	cmd := exec.Command("go", "tool", "cover", "-func="+coverageOutPath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run go tool cover: %w", err)
	}

	// Parse the output and convert to LCOV format
	lines := strings.Split(string(output), "\n")
	var lcovLines []string
	var currentFile string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total:") {
			continue
		}

		// Parse format: github.com/swantron/difftron/internal/hunk/parser.go:42:	ParseGitDiff	100.0%
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		fileAndLine := parts[0]
		fileParts := strings.Split(fileAndLine, ":")
		if len(fileParts) < 2 {
			continue
		}

		filePath := fileParts[0]
		// Remove the module path prefix
		filePath = strings.TrimPrefix(filePath, "github.com/swantron/difftron/")
		filePath = strings.TrimPrefix(filePath, "github.com\\swantron\\difftron\\") // Windows

		// Get coverage percentage
		coverageStr := parts[len(parts)-1]
		coverageStr = strings.TrimSuffix(coverageStr, "%")
		var coverage float64
		fmt.Sscanf(coverageStr, "%f", &coverage)

		// Start new file record
		if filePath != currentFile {
			if currentFile != "" {
				lcovLines = append(lcovLines, "end_of_record")
			}
			lcovLines = append(lcovLines, "SF:"+filePath)
			currentFile = filePath
		}

		// For Go coverage, we mark lines as covered if coverage > 0
		// Note: This is simplified - Go's coverage.out has line-by-line data
		// but we're using function-level coverage here
		// For true line-by-line, we'd need to parse coverage.out binary format
		if coverage > 0 {
			// Mark all lines in this function as covered
			// This is approximate - for exact line coverage, parse coverage.out directly
			lcovLines = append(lcovLines, fmt.Sprintf("DA:%s,1", fileParts[1]))
		}
	}

	if currentFile != "" {
		lcovLines = append(lcovLines, "end_of_record")
	}

	// Write LCOV file
	return os.WriteFile(outputPath, []byte(strings.Join(lcovLines, "\n")), 0644)
}

// ParseGoCoverage parses Go's native coverage.out format directly
// This is more accurate than converting to LCOV
func ParseGoCoverage(coverageOutPath string) (*Report, error) {
	// Use go tool cover to get line-by-line coverage
	cmd := exec.Command("go", "tool", "cover", "-func="+coverageOutPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go coverage: %w", err)
	}

	report := &Report{
		FileCoverage: make(map[string]*CoverageData),
	}

	lines := strings.Split(string(output), "\n")
	var currentFile string
	var currentCoverage *CoverageData

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "total:") {
			continue
		}

		// Parse: github.com/swantron/difftron/internal/hunk/parser.go:42:	ParseGitDiff	100.0%
		// Extract file path and line number
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		filePath := line[:colonIdx]
		// Normalize file path
		filePath = strings.TrimPrefix(filePath, "github.com/swantron/difftron/")
		filePath = strings.TrimPrefix(filePath, "github.com\\swantron\\difftron\\")

		// Extract line number
		rest := line[colonIdx+1:]
		lineNumEnd := strings.Index(rest, ":")
		if lineNumEnd == -1 {
			continue
		}

		var lineNum int
		fmt.Sscanf(rest[:lineNumEnd], "%d", &lineNum)

		// Extract coverage percentage
		parts := strings.Fields(rest)
		if len(parts) < 2 {
			continue
		}
		coverageStr := parts[len(parts)-1]
		coverageStr = strings.TrimSuffix(coverageStr, "%")
		var coverage float64
		fmt.Sscanf(coverageStr, "%f", &coverage)

		// Initialize file coverage if needed
		if filePath != currentFile {
			currentFile = filePath
			currentCoverage = &CoverageData{
				LineHits: make(map[int]int),
			}
			report.FileCoverage[filePath] = currentCoverage
		}

		// Mark line as covered if coverage > 0
		// Note: This is function-level, not line-level
		// For true line coverage, we'd need to parse the binary coverage.out format
		if coverage > 0 {
			currentCoverage.LineHits[lineNum] = 1
			currentCoverage.TotalLines++
			currentCoverage.CoveredLines++
		}
	}

	return report, nil
}

// DetectCoverageFormat detects if a coverage file is LCOV or Go format
func DetectCoverageFormat(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Read first few bytes to check format without reading entire file
	content := string(data)
	if len(content) > 1000 {
		content = content[:1000] // Only check first 1KB
	}

	trimmed := strings.TrimSpace(content)

	// Check for Cobertura XML format
	if strings.HasPrefix(trimmed, "<?xml") || strings.Contains(content, "<coverage") {
		// Check if it's Cobertura format
		if strings.Contains(content, "cobertura") || strings.Contains(content, "coverage") {
			// Verify it has Cobertura-specific elements
			if strings.Contains(content, "<package") || strings.Contains(content, "<class") {
				return "cobertura", nil
			}
		}
	}

	// Check for LCOV format markers (must be first)
	if strings.HasPrefix(trimmed, "TN:") ||
		strings.HasPrefix(trimmed, "SF:") ||
		(strings.Contains(content, "SF:") && strings.Contains(content, "DA:")) {
		return "lcov", nil
	}

	// Check for Go coverage format
	// Go coverage.out files start with "mode:" on first line
	if strings.HasPrefix(trimmed, "mode:") {
		return "go", nil
	}

	// If file extension is .out and starts with binary-looking data or "mode:", assume Go
	if filepath.Ext(filePath) == ".out" {
		// Check if it looks like Go coverage (starts with "mode:" or is binary)
		if strings.HasPrefix(trimmed, "mode:") {
			return "go", nil
		}
		// If it's not text (doesn't contain common text markers), assume binary Go format
		if !strings.Contains(content, "SF:") && !strings.Contains(content, "TN:") {
			return "go", nil
		}
	}

	// Default to LCOV
	return "lcov", nil
}
