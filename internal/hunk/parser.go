package hunk

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// Hunk represents a changed line in a file
type Hunk struct {
	File      string
	Line      int
	IsAdded   bool
	IsRemoved bool
}

// ParseResult contains the parsed diff information
type ParseResult struct {
	// ChangedLines maps file path -> line number -> true if changed
	ChangedLines map[string]map[int]bool
	// AddedLines maps file path -> line number -> true if added
	AddedLines map[string]map[int]bool
	// RemovedLines maps file path -> line number -> true if removed
	RemovedLines map[string]map[int]bool
	// NewFiles tracks which files are new (didn't exist in base)
	NewFiles map[string]bool
	// ModifiedFiles tracks which files existed in base and were modified
	ModifiedFiles map[string]bool
}

// ParseGitDiff parses git diff output and returns a map of changed lines
// The output format is: map[filepath]map[lineNumber]bool
func ParseGitDiff(diffOutput string) (*ParseResult, error) {
	result := &ParseResult{
		ChangedLines:  make(map[string]map[int]bool),
		AddedLines:    make(map[string]map[int]bool),
		RemovedLines:  make(map[string]map[int]bool),
		NewFiles:      make(map[string]bool),
		ModifiedFiles: make(map[string]bool),
	}

	scanner := bufio.NewScanner(strings.NewReader(diffOutput))
	var currentFile string
	var currentFileOldPath string // Track the old path to detect new files
	var currentLine int           // Line number in the new file version

	for scanner.Scan() {
		line := scanner.Text()

		// Track the old file path
		// Format: --- a/path/to/file.go
		if strings.HasPrefix(line, "--- a/") {
			currentFileOldPath = strings.TrimPrefix(line, "--- a/")
			continue
		}

		// Track the file being modified
		// Format: +++ b/path/to/file.go
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			if currentFile == "/dev/null" {
				// File was deleted, skip
				currentFile = ""
				currentFileOldPath = ""
				continue
			}
			
			result.ChangedLines[currentFile] = make(map[int]bool)
			result.AddedLines[currentFile] = make(map[int]bool)
			result.RemovedLines[currentFile] = make(map[int]bool)
			
			// Detect if this is a new file
			// New files have old path as /dev/null or empty
			if currentFileOldPath == "/dev/null" || currentFileOldPath == "" {
				result.NewFiles[currentFile] = true
			} else {
				result.ModifiedFiles[currentFile] = true
			}
			
			currentFileOldPath = "" // Reset for next file
			continue
		}

		// Parse hunk header
		// Format: @@ -oldStart,oldCount +newStart,newCount @@
		// Example: @@ -10,5 +15,7 @@
		if strings.HasPrefix(line, "@@") {
			parts := strings.Fields(line)
			if len(parts) < 3 {
				continue
			}

			// Extract the new file line number
			newFilePart := parts[2]
			if !strings.HasPrefix(newFilePart, "+") {
				continue
			}

			// Handle both formats: +15,7 and +15
			newFilePart = strings.TrimPrefix(newFilePart, "+")
			lineParts := strings.Split(newFilePart, ",")
			startLine, err := strconv.Atoi(lineParts[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse line number in hunk header: %w", err)
			}

			// Line numbers in git diff are 1-indexed
			// The startLine is the first line number shown in the hunk
			// We'll increment before processing each line, so start one before
			currentLine = startLine - 1
			continue
		}

		// Skip if we don't have a current file
		if currentFile == "" {
			continue
		}

		// Process diff lines
		// Note: We increment the line counter BEFORE processing, so the first
		// line after a hunk header gets the correct line number
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			// Added line
			currentLine++
			result.ChangedLines[currentFile][currentLine] = true
			result.AddedLines[currentFile][currentLine] = true
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			// Removed line (tracked for context, but not counted in new file)
			// Don't increment currentLine for removed lines in the new file
		} else if strings.HasPrefix(line, " ") || (!strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-")) {
			// Context line (unchanged) - starts with space or is not a +/- line
			// Increment line counter for context lines
			currentLine++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading diff: %w", err)
	}

	return result, nil
}

// GetChangedFiles returns a list of all files that have changes
func (r *ParseResult) GetChangedFiles() []string {
	files := make([]string, 0, len(r.ChangedLines))
	for file := range r.ChangedLines {
		files = append(files, file)
	}
	return files
}

// GetChangedLinesForFile returns the set of changed line numbers for a file
func (r *ParseResult) GetChangedLinesForFile(file string) map[int]bool {
	return r.ChangedLines[file]
}

// GetAddedLinesForFile returns the set of added line numbers for a file
func (r *ParseResult) GetAddedLinesForFile(file string) map[int]bool {
	return r.AddedLines[file]
}

// IsNewFile returns true if the file is new (didn't exist in base)
func (r *ParseResult) IsNewFile(file string) bool {
	return r.NewFiles[file]
}

// IsModifiedFile returns true if the file existed in base and was modified
func (r *ParseResult) IsModifiedFile(file string) bool {
	return r.ModifiedFiles[file]
}

// HasChanges returns true if there are any changes
func (r *ParseResult) HasChanges() bool {
	return len(r.ChangedLines) > 0
}
