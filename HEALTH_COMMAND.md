# Health Command Documentation

The `difftron health` command provides comprehensive testing health analysis by aggregating coverage across multiple test types and comparing against baseline coverage.

## Features

- **Multi-test-type aggregation**: Combines coverage from unit, API, and functional tests
- **Baseline comparison**: Detects coverage regressions by comparing against baseline
- **New vs modified file analysis**: Separately tracks coverage for new and modified files
- **Actionable insights**: Generates prioritized recommendations
- **PR/MR integration**: Can post comments on GitHub PRs and GitLab MRs
- **Multiple output formats**: JSON, Markdown, and structured text

## Usage

### Basic Usage

```bash
# Single test type
difftron health --unit-coverage coverage.out --threshold 80

# Multiple test types
difftron health \
  --unit-coverage unit-coverage.out \
  --api-coverage api-coverage.info \
  --functional-coverage functional-coverage.info \
  --threshold 80
```

### With Baseline Comparison

```bash
difftron health \
  --unit-coverage unit-coverage.out \
  --baseline-unit-coverage baseline-coverage.out \
  --threshold 80
```

### Separate Thresholds

```bash
difftron health \
  --unit-coverage unit-coverage.out \
  --threshold 80 \
  --threshold-new 90 \
  --threshold-modified 75
```

### Output Formats

```bash
# JSON output
difftron health --unit-coverage coverage.out --output json > health.json

# Markdown output
difftron health --unit-coverage coverage.out --output markdown > health.md

# Structured text (default)
difftron health --unit-coverage coverage.out --output text
```

### PR/MR Comments

```bash
# GitHub PR (requires GITHUB_TOKEN)
export GITHUB_TOKEN=your_token
export GITHUB_PR_NUMBER=123
difftron health --unit-coverage coverage.out --comment-pr

# GitLab MR (requires GITLAB_TOKEN)
export GITLAB_TOKEN=your_token
export CI_MERGE_REQUEST_IID=456
difftron health --unit-coverage coverage.out --comment-mr
```

## Command Options

| Option | Description | Default |
|--------|-------------|---------|
| `--unit-coverage` | Path to unit test coverage file | Required (at least one) |
| `--api-coverage` | Path to API test coverage file | - |
| `--functional-coverage` | Path to functional test coverage file | - |
| `--baseline-unit-coverage` | Path to baseline unit test coverage | - |
| `--baseline-api-coverage` | Path to baseline API test coverage | - |
| `--baseline-functional-coverage` | Path to baseline functional test coverage | - |
| `--threshold` | Coverage threshold percentage | 80.0 |
| `--threshold-new` | Coverage threshold for new files | Same as threshold |
| `--threshold-modified` | Coverage threshold for modified files | Same as threshold |
| `--output` | Output format (text, json, markdown) | text |
| `--output-file` | Output file path | stdout |
| `--base` | Base git ref for diff | Auto-detect |
| `--head` | Head git ref for diff | Auto-detect |
| `--comment-pr` | Post comment on GitHub PR | false |
| `--comment-mr` | Post comment on GitLab MR | false |

## Output Format

### JSON Output

```json
{
  "overall_coverage": 85.5,
  "changed_coverage": 82.3,
  "total_files": 10,
  "changed_files": 3,
  "healthy_files": 2,
  "at_risk_files": 1,
  "regressing_files": 0,
  "file_health": {
    "path/to/file.go": {
      "status": "healthy",
      "is_new_file": false,
      "overall_coverage": 90.0,
      "changed_coverage": 85.0,
      "baseline_coverage": 80.0,
      "coverage_delta": 5.0,
      "changed_lines": 20,
      "covered_lines": 17,
      "uncovered_lines": 3,
      "unit_test_coverage": 85.0,
      "api_test_coverage": 0.0,
      "functional_test_coverage": 0.0
    }
  },
  "insights": [...],
  "recommendations": [...]
}
```

### Markdown Output

The markdown output provides a comprehensive report with:
- Summary statistics
- Test type breakdown
- File-level health details
- Insights and recommendations

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run difftron health
  run: |
    difftron health \
      --unit-coverage unit-coverage.out \
      --api-coverage api-coverage.info \
      --threshold 80 \
      --output json > health-report.json \
      --comment-pr
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### GitLab CI

```yaml
difftron-health:
  script:
    - difftron health \
        --unit-coverage unit-coverage.out \
        --threshold 80 \
        --output markdown \
        --comment-mr
  variables:
    GITLAB_TOKEN: $CI_JOB_TOKEN
```

## Exit Codes

- `0`: Health check passed (all thresholds met, no regressions)
- `1`: Health check failed (thresholds not met or regressions detected)
- `2`: Error occurred during analysis

## Examples

See [HOLISTIC_HEALTH.md](./HOLISTIC_HEALTH.md) for more detailed examples and use cases.
