# Baseline Coverage Tracking

## Problem

Traditional coverage tools have a significant flaw: when you touch a file that didn't have proper coverage before, the overall coverage drops because:

1. **New files** are being tracked for the first time
2. **Untested code** in modified files is now visible
3. The coverage tool sees **all lines** in the file, not just what changed

This penalizes developers unfairly - they're being blamed for untested code that existed before their changes.

## Solution

Difftron now tracks **baseline coverage** and separates metrics for:

1. **New Files** - Files that didn't exist in the base branch
2. **Modified Files** - Files that existed and were changed

### Key Features

#### 1. New File Detection
- Automatically detects new files vs modified files from git diff
- New files have `--- /dev/null` in git diff
- Modified files have actual file paths in both old and new versions

#### 2. Baseline Coverage Tracking
- For modified files, tracks what the coverage was **before** your changes
- Only counts coverage for the **lines that changed**, not the entire file
- Calculates baseline coverage percentage for changed lines

#### 3. Separate Metrics

The `AnalysisResult` now includes:

```go
type AnalysisResult struct {
    // Overall metrics (all files)
    TotalChangedLines int
    CoveredLines int
    UncoveredLines int
    CoveragePercentage float64
    
    // New file metrics (only new files)
    NewFileMetrics *FileTypeMetrics
    
    // Modified file metrics (only modified files)
    ModifiedFileMetrics *FileTypeMetrics
}
```

#### 4. Per-File Baseline Tracking

Each `FileResult` includes:

```go
type FileResult struct {
    // ... existing fields ...
    
    IsNewFile bool  // true if file didn't exist in base
    BaselineCoveragePercentage float64  // coverage before changes (for modified files)
}
```

## Usage

### Basic Analysis (No Baseline)

```go
result, err := analyzer.Analyze(diffResult, coverageReport)
```

This works as before, but now tracks new vs modified files.

### Analysis With Baseline

```go
// Get baseline coverage from previous commit or CI artifact
baselineReport, err := coverage.ParseLCOV("baseline-coverage.info")

// Analyze with baseline
result, err := analyzer.AnalyzeWithBaseline(diffResult, currentReport, baselineReport)
```

### Accessing Metrics

```go
// Overall coverage (all files)
overallCoverage := result.CoveragePercentage

// New file coverage only
newFileCoverage := result.NewFileMetrics.CoveragePercentage

// Modified file coverage only
modifiedFileCoverage := result.ModifiedFileMetrics.CoveragePercentage

// Per-file baseline
for filePath, fileResult := range result.FileResults {
    if fileResult.IsNewFile {
        fmt.Printf("%s is NEW - coverage: %.1f%%\n", filePath, fileResult.CoveragePercentage)
    } else {
        fmt.Printf("%s is MODIFIED - coverage: %.1f%% (baseline: %.1f%%)\n", 
            filePath, fileResult.CoveragePercentage, fileResult.BaselineCoveragePercentage)
    }
}
```

## Benefits

1. **Fair Assessment**: Only penalize developers for code they actually changed
2. **Accurate Metrics**: See "actual" coverage vs "reported" coverage
3. **Better Insights**: Understand if coverage dropped or if you're just seeing untested code for the first time
4. **Separate Thresholds**: Can set different thresholds for new files vs modified files

## Example Scenarios

### Scenario 1: New File with Good Coverage

```
New file: utils.go
- 100 lines added
- 95 lines covered
- Coverage: 95%
```

**Result**: ✅ Good coverage for new file, doesn't affect overall project coverage negatively

### Scenario 2: Modified File with Baseline

```
Modified file: legacy.go
- 10 lines changed
- Current coverage: 20% (2/10 lines covered)
- Baseline coverage: 0% (0/10 lines were covered before)
```

**Result**: ✅ Coverage improved from 0% to 20% - this is progress, not a problem!

### Scenario 3: Modified File Coverage Drop

```
Modified file: api.go
- 5 lines changed
- Current coverage: 40% (2/5 lines covered)
- Baseline coverage: 80% (4/5 lines were covered before)
```

**Result**: ⚠️ Coverage dropped from 80% to 40% - this is a real regression

## Implementation Details

### Git Diff Parsing

The parser now tracks:
- `--- a/file.go` (old file path)
- `+++ b/file.go` (new file path)
- `--- /dev/null` indicates a new file
- `+++ /dev/null` indicates a deleted file

### Baseline Coverage Calculation

For modified files, baseline coverage is calculated as:
- Count how many of the **changed lines** were covered in baseline
- Divide by total changed lines
- This gives baseline coverage percentage for the delta

### Coverage Accuracy

The key insight: **We only count coverage for lines that changed**, not the entire file. This gives you the "actual" coverage of your changes, not the "reported" coverage of the whole file.

## Future Enhancements

1. **CI Integration**: Automatically fetch baseline coverage from previous CI runs
2. **Git History**: Calculate baseline from git history automatically
3. **Thresholds**: Separate thresholds for new files vs modified files
4. **Trends**: Track coverage trends over time per file
5. **Warnings**: Warn when touching files with low baseline coverage
