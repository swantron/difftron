# Dogfooding Difftron

This guide shows how to use Difftron on itself to ensure code quality.

## Quick Start

```bash
# Analyze your last commit
go run scripts/task.go dogfood

# Analyze specific commit range
go run scripts/task.go dogfood HEAD~5 HEAD

# Or use the dogfood script directly
./scripts/dogfood.sh
```

## How It Works

1. **Generates Coverage**: Runs `go test -coverprofile=coverage.out` to get coverage data
2. **Gets Diff**: Uses `git diff` to see what changed
3. **Analyzes**: Runs difftron on the diff against the coverage

## GitHub Actions Integration

The `.github/workflows/difftron.yml` workflow automatically runs on:
- **Pull Requests** to `main` - Posts coverage analysis as a PR comment
- **Pushes to main** - Runs as a status check

### Workflow Features

- ✅ Generates coverage for all Go packages
- ✅ Analyzes diff between base and head commits
- ✅ Posts PR comments with coverage report
- ✅ Sets status checks on pushes
- ✅ Uploads coverage artifacts

## Manual Testing

### Before Committing

```bash
# Make your changes, then:

# 1. Generate coverage
go test -coverprofile=coverage.out ./...

# 2. See what you changed
git diff HEAD > my-changes.diff

# 3. Analyze
./bin/difftron analyze \
  --coverage coverage.out \
  --diff my-changes.diff \
  --threshold 80
```

### Testing Specific Commits

```bash
# Compare two commits
git diff abc123..def456 > changes.diff
./bin/difftron analyze \
  --coverage coverage.out \
  --diff changes.diff
```

## Coverage Format Support

Difftron supports:
- **LCOV format** (`.info` files) - Full line-by-line coverage (recommended)
- **Go coverage format** (`.out` files) - Function-level coverage (approximate)

**Note**: Go's native coverage format provides function-level coverage when parsed via `go tool cover -func`. For more accurate line-by-line analysis, consider converting to LCOV format using tools like `gocov` or `gocovmerge`.

When you use `go test -coverprofile=coverage.out`, difftron automatically detects and parses the Go format, but results may be approximate since it uses function-level data.

## CI/CD Integration

### GitHub Actions

The workflow is already set up in `.github/workflows/difftron.yml`. It will:
1. Run on every PR and push to main
2. Generate coverage
3. Analyze changes
4. Post results

### Customization

Edit `.github/workflows/difftron.yml` to:
- Change the coverage threshold (default: 80%)
- Modify when it runs (currently: PRs and pushes to main)
- Adjust the comment format
- Add more checks

### GitLab CI

To add GitLab CI support, create `.gitlab-ci.yml`:

```yaml
difftron:
  image: golang:1.21
  script:
    - go test -coverprofile=coverage.out ./...
    - go build -o bin/difftron ./cmd/difftron
    - git diff $CI_MERGE_REQUEST_DIFF_BASE_SHA..$CI_COMMIT_SHA > diff.patch
    - ./bin/difftron analyze --coverage coverage.out --diff diff.patch --threshold 80
  only:
    - merge_requests
```

## Troubleshooting

### No Coverage Data

If you see "No coverage data found":
- Make sure you ran `go test -coverprofile=coverage.out`
- Check that coverage.out exists and has content
- Verify you're analyzing files that have tests

### Diff Shows No Changes

If difftron says "No changes detected":
- Check that your git diff has content: `git diff HEAD~1 HEAD`
- Verify you're comparing the right commits
- Make sure you're in a git repository

### Coverage Format Issues

If you get format errors:
- Go format: Use `go test -coverprofile=coverage.out`
- LCOV format: Ensure file starts with `TN:` or `SF:`
- Check file extension: `.out` for Go, `.info` for LCOV

## Best Practices

1. **Run Before Committing**: Use `go run scripts/task.go dogfood` before pushing
2. **Set Reasonable Thresholds**: Start with 80%, adjust based on your project
3. **Fix Uncovered Lines**: Add tests for uncovered code paths
4. **Monitor Trends**: Track coverage over time in CI/CD

## Example Output

```
Dogfooding Difftron on itself
==================================
Base: HEAD~1
Head: HEAD

Building difftron...
Generating coverage...
Generating diff...

Running difftron analysis...

Difftron Coverage Analysis
==========================

Overall Coverage: 85.7% (6/7 lines covered)

✓ Coverage threshold met (85.7% >= 80.0%)

Per-File Results:
-----------------

internal/analyzer/analyzer.go
  Coverage: 100.0% (3/3 lines covered)

internal/hunk/parser.go
  Coverage: 75.0% (3/4 lines covered)
  Uncovered lines: [42]
```

## Next Steps

- Add more test cases to improve coverage
- Integrate into your CI/CD pipeline
- Set up coverage tracking over time
- Add coverage badges to README
