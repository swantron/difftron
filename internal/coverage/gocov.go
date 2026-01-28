package coverage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
// Supports both text format (mode: set/count) and binary format
// For text format, parses line-by-line coverage ranges
func ParseGoCoverage(coverageOutPath string) (*Report, error) {
	// Try to parse as text format first
	report, err := parseGoCoverageText(coverageOutPath)
	if err == nil {
		return report, nil
	}

	// If text parsing fails, fall back to function-level parsing via go tool cover
	return parseGoCoverageFunc(coverageOutPath)
}

// parseGoCoverageText parses Go coverage.out text format (mode: set or mode: count)
// Format: mode: set
//         file:startLine.startCol,endLine.endCol count statements
func parseGoCoverageText(coverageOutPath string) (*Report, error) {
	file, err := os.Open(coverageOutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open coverage file: %w", err)
	}
	defer file.Close()

	report := &Report{
		FileCoverage: make(map[string]*CoverageData),
	}

	scanner := bufio.NewScanner(file)
	var mode string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse mode line: "mode: set" or "mode: count"
		if strings.HasPrefix(line, "mode:") {
			mode = strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
			continue
		}

		// Parse coverage line: file:startLine.startCol,endLine.endCol count statements
		// Example: github.com/swantron/difftron/internal/hunk/parser.go:42.0,43.0 1 0
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		// Parse file and line range
		fileAndRange := parts[0]
		colonIdx := strings.LastIndex(fileAndRange, ":")
		if colonIdx == -1 {
			continue
		}

		filePath := fileAndRange[:colonIdx]
		rangeStr := fileAndRange[colonIdx+1:]

		// Parse range: startLine.startCol,endLine.endCol
		commaIdx := strings.Index(rangeStr, ",")
		if commaIdx == -1 {
			continue
		}

		startStr := rangeStr[:commaIdx]
		endStr := rangeStr[commaIdx+1:]

		// Extract line numbers (ignore column numbers)
		startDotIdx := strings.Index(startStr, ".")
		if startDotIdx != -1 {
			startStr = startStr[:startDotIdx]
		}
		endDotIdx := strings.Index(endStr, ".")
		if endDotIdx != -1 {
			endStr = endStr[:endDotIdx]
		}

		startLine, err := strconv.Atoi(startStr)
		if err != nil {
			continue
		}
		endLine, err := strconv.Atoi(endStr)
		if err != nil {
			continue
		}

		// Parse count (number of times this range was executed)
		count, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		// Normalize file path (remove module prefix)
		filePath = normalizeGoFilePath(filePath)

		// Get or create file coverage
		fileCoverage := report.FileCoverage[filePath]
		if fileCoverage == nil {
			fileCoverage = &CoverageData{
				LineHits: make(map[int]int),
			}
			report.FileCoverage[filePath] = fileCoverage
		}

		// Mark lines in range as covered (if count > 0)
		// For mode: set, count is 0 or 1
		// For mode: count, count is the actual execution count
		for line := startLine; line <= endLine; line++ {
			// Track if this line was already seen (for TotalLines counting)
			wasAlreadySeen := fileCoverage.LineHits[line] > 0

			if count > 0 {
				// Update hit count (take maximum if already set, or sum for count mode)
				if existingCount, exists := fileCoverage.LineHits[line]; exists {
					// If mode is count, we might want to sum, but typically we take max
					// For now, take maximum to avoid double counting
					if count > existingCount {
						fileCoverage.LineHits[line] = count
					}
				} else {
					fileCoverage.LineHits[line] = count
					// New line covered
					fileCoverage.CoveredLines++
				}
			}

			// Count total lines (only once per line)
			if !wasAlreadySeen {
				fileCoverage.TotalLines++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading coverage file: %w", err)
	}

	// If we didn't see a mode line, this might not be a text format file
	if mode == "" && len(report.FileCoverage) == 0 {
		return nil, fmt.Errorf("not a valid Go coverage text format (no mode line found)")
	}

	return report, nil
}


// normalizeGoFilePath normalizes Go file paths by removing module prefixes
func normalizeGoFilePath(filePath string) string {
	// Convert Windows paths to forward slashes
	filePath = filepath.ToSlash(filePath)

	// Try to detect and remove common module prefixes
	// This is a heuristic - actual module path detection would require go.mod parsing
	commonPrefixes := []string{
		"github.com/swantron/difftron/",
		"github.com\\swantron\\difftron\\",
	}

	for _, prefix := range commonPrefixes {
		if strings.HasPrefix(filePath, prefix) {
			return strings.TrimPrefix(filePath, prefix)
		}
	}

	return filePath
}

// parseGoCoverageFunc parses Go coverage using go tool cover -func (function-level fallback)
func parseGoCoverageFunc(coverageOutPath string) (*Report, error) {
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
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		filePath := line[:colonIdx]
		filePath = normalizeGoFilePath(filePath)

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
		// Note: This is function-level approximation
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
