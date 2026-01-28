package coverage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// CoverageData represents coverage information for a file
type CoverageData struct {
	// LineHits maps line number -> hit count
	LineHits map[int]int
	// TotalLines is the total number of executable lines
	TotalLines int
	// CoveredLines is the number of lines with hits > 0
	CoveredLines int
}

// Report contains coverage data for multiple files
type Report struct {
	// FileCoverage maps file path -> CoverageData
	FileCoverage map[string]*CoverageData
}

// ParseLCOV parses an LCOV format coverage file (.info)
// Returns a Report containing coverage data for all files
func ParseLCOV(filePath string) (*Report, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open LCOV file: %w", err)
	}
	defer file.Close()

	report := &Report{
		FileCoverage: make(map[string]*CoverageData),
	}

	scanner := bufio.NewScanner(file)
	var currentFile string
	var currentCoverage *CoverageData

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// SF: Source file
		// Format: SF:/path/to/file.go
		if strings.HasPrefix(line, "SF:") {
			currentFile = line[3:]
			currentCoverage = &CoverageData{
				LineHits: make(map[int]int),
			}
			report.FileCoverage[currentFile] = currentCoverage
			continue
		}

		// Skip if we don't have a current file
		if currentFile == "" || currentCoverage == nil {
			continue
		}

		// DA: Line data
		// Format: DA:line_number,hit_count
		if strings.HasPrefix(line, "DA:") {
			data := strings.TrimPrefix(line, "DA:")
			parts := strings.Split(data, ",")
			if len(parts) != 2 {
				continue
			}

			lineNum, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}

			hits, err := strconv.Atoi(parts[1])
			if err != nil {
				continue
			}

			currentCoverage.LineHits[lineNum] = hits
			currentCoverage.TotalLines++

			if hits > 0 {
				currentCoverage.CoveredLines++
			}
			continue
		}

		// end_of_record marks the end of a file's coverage data
		if line == "end_of_record" {
			currentFile = ""
			currentCoverage = nil
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading LCOV file: %w", err)
	}

	return report, nil
}

// GetCoverageForFile returns coverage data for a specific file
// Returns nil if the file is not in the report
func (r *Report) GetCoverageForFile(filePath string) *CoverageData {
	return r.FileCoverage[filePath]
}

// GetCoverageForLine returns the hit count for a specific line in a file
// Returns 0 if the line is not covered or file doesn't exist
func (r *Report) GetCoverageForLine(filePath string, lineNum int) int {
	coverage := r.GetCoverageForFile(filePath)
	if coverage == nil {
		return 0
	}
	return coverage.LineHits[lineNum]
}

// IsLineCovered returns true if a line has been executed (hits > 0)
func (r *Report) IsLineCovered(filePath string, lineNum int) bool {
	return r.GetCoverageForLine(filePath, lineNum) > 0
}

var (
	repoRootCache     string
	repoRootCacheOnce sync.Once
)

// getRepoRoot detects the git repository root directory
func getRepoRoot() string {
	repoRootCacheOnce.Do(func() {
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		output, err := cmd.Output()
		if err == nil {
			repoRootCache = strings.TrimSpace(string(output))
		}
	})
	return repoRootCache
}

// NormalizePath attempts to normalize file paths for comparison
// This helps match git diff paths with LCOV file paths
// It handles:
// - Windows/Unix path separators (using filepath.ToSlash)
// - Absolute vs relative paths
// - Repo root rebasing
// - Leading "./" and "/" prefixes
func NormalizePath(path string) string {
	if path == "" {
		return path
	}

	// Convert to forward slashes for cross-platform consistency
	normalized := filepath.ToSlash(path)

	// Remove leading "./" if present
	normalized = strings.TrimPrefix(normalized, "./")

	// Try to rebase relative to repo root if it's an absolute path
	if filepath.IsAbs(path) {
		repoRoot := getRepoRoot()
		if repoRoot != "" {
			repoRootSlash := filepath.ToSlash(repoRoot)
			if strings.HasPrefix(normalized, repoRootSlash+"/") {
				normalized = strings.TrimPrefix(normalized, repoRootSlash+"/")
			} else if normalized == repoRootSlash {
				normalized = ""
			} else {
				// Absolute path outside repo root - remove leading "/"
				normalized = strings.TrimPrefix(normalized, "/")
			}
		} else {
			// If we can't detect repo root, just remove leading "/"
			normalized = strings.TrimPrefix(normalized, "/")
		}
	} else {
		// Remove leading "/" if present (for relative paths that start with /)
		normalized = strings.TrimPrefix(normalized, "/")
	}

	return normalized
}

// FindMatchingPath attempts to find a matching path in the coverage report
// by trying multiple normalization strategies
func FindMatchingPath(targetPath string, availablePaths map[string]*CoverageData) string {
	// Strategy 1: Exact match
	if _, exists := availablePaths[targetPath]; exists {
		return targetPath
	}

	// Strategy 2: Normalized match
	normalized := NormalizePath(targetPath)
	if normalized != targetPath {
		if _, exists := availablePaths[normalized]; exists {
			return normalized
		}
	}

	// Strategy 3: Try all available paths with normalization
	for availablePath := range availablePaths {
		if NormalizePath(availablePath) == normalized {
			return availablePath
		}
	}

	// Strategy 4: Stem-only match (filename only) as last resort
	stem := filepath.Base(targetPath)
	for availablePath := range availablePaths {
		if filepath.Base(availablePath) == stem {
			return availablePath
		}
	}

	// No match found
	return ""
}
