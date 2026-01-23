package hunk

import (
	"testing"
)

func TestParseGitDiff(t *testing.T) {
	tests := []struct {
		name        string
		diffOutput  string
		wantFiles   []string
		wantLines   map[string][]int // file -> expected line numbers
		wantAdded   map[string][]int // file -> expected added line numbers
		expectError bool
	}{
		{
			name: "simple addition",
			diffOutput: `diff --git a/file.go b/file.go
index 1234567..abcdefg 100644
--- a/file.go
+++ b/file.go
@@ -5,3 +5,5 @@ func main() {
 	fmt.Println("hello")
+	fmt.Println("new line 1")
+	fmt.Println("new line 2")
 	fmt.Println("world")
`,
			wantFiles: []string{"file.go"},
			wantLines: map[string][]int{
				"file.go": {6, 7},
			},
			wantAdded: map[string][]int{
				"file.go": {6, 7},
			},
			expectError: false,
		},
		{
			name: "multiple files",
			diffOutput: `diff --git a/file1.go b/file1.go
index 111..222 100644
--- a/file1.go
+++ b/file1.go
@@ -10,2 +10,3 @@
 	oldLine
+	newLine
 	oldLine2

diff --git a/file2.go b/file2.go
index 333..444 100644
--- a/file2.go
+++ b/file2.go
@@ -5,1 +5,2 @@
 	oldLine
+	newLine
`,
			wantFiles: []string{"file1.go", "file2.go"},
			wantLines: map[string][]int{
				"file1.go": {11},
				"file2.go": {6},
			},
			wantAdded: map[string][]int{
				"file1.go": {11},
				"file2.go": {6},
			},
			expectError: false,
		},
		{
			name: "empty diff",
			diffOutput: `diff --git a/file.go b/file.go
index 123..123 100644
`,
			wantFiles: []string{},
			wantLines: map[string][]int{},
			wantAdded: map[string][]int{},
			expectError: false,
		},
		{
			name: "hunk with context",
			diffOutput: `diff --git a/file.go b/file.go
index 123..456 100644
--- a/file.go
+++ b/file.go
@@ -10,5 +10,7 @@ func test() {
 	line1
 	line2
+	new1
+	new2
 	line3
 	line4
`,
			wantFiles: []string{"file.go"},
			// The hunk header +10,7 means starting at line 10
			// line1 is line 10, line2 is line 11, new1 is line 12, new2 is line 13
			wantLines: map[string][]int{
				"file.go": {12, 13},
			},
			wantAdded: map[string][]int{
				"file.go": {12, 13},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseGitDiff(tt.diffOutput)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check files
			gotFiles := result.GetChangedFiles()
			if len(gotFiles) != len(tt.wantFiles) {
				t.Errorf("expected %d files, got %d: %v", len(tt.wantFiles), len(gotFiles), gotFiles)
			}

			// Check lines for each file
			for file, wantLines := range tt.wantLines {
				gotLines := result.GetChangedLinesForFile(file)
				if len(gotLines) != len(wantLines) {
					t.Errorf("file %s: expected %d changed lines, got %d", file, len(wantLines), len(gotLines))
					continue
				}

				for _, lineNum := range wantLines {
					if !gotLines[lineNum] {
						t.Errorf("file %s: expected line %d to be changed", file, lineNum)
					}
				}
			}

			// Check added lines
			for file, wantAdded := range tt.wantAdded {
				gotAdded := result.GetAddedLinesForFile(file)
				if len(gotAdded) != len(wantAdded) {
					t.Errorf("file %s: expected %d added lines, got %d", file, len(wantAdded), len(gotAdded))
					continue
				}

				for _, lineNum := range wantAdded {
					if !gotAdded[lineNum] {
						t.Errorf("file %s: expected line %d to be added", file, lineNum)
					}
				}
			}
		})
	}
}

func TestParseResult_HasChanges(t *testing.T) {
	result := &ParseResult{
		ChangedLines: make(map[string]map[int]bool),
		AddedLines:   make(map[string]map[int]bool),
		RemovedLines: make(map[string]map[int]bool),
	}

	if result.HasChanges() {
		t.Error("expected no changes for empty result")
	}

	result.ChangedLines["file.go"] = map[int]bool{1: true}
	if !result.HasChanges() {
		t.Error("expected changes to be detected")
	}
}
