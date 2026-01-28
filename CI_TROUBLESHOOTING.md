# CI Troubleshooting Guide

Common CI issues and their solutions.

## Issue: Cache errors ("Cannot open: File exists", tar failures)

**Symptom**: Cache-related errors like "Cannot open: File exists" or "tar failed with exit code 2"

**Cause**: Manual cache actions can conflict with filesystem state or have permission issues.

**Solution**: The workflow now uses `setup-go@v5`'s built-in caching via `cache-dependency-path: go.sum`. This is more reliable and handles cache management automatically.

**If you see cache errors**:
1. Check that `setup-go@v5` is being used (not an older version)
2. Ensure `cache-dependency-path: go.sum` is set
3. The built-in cache should work without manual cache steps

**Note**: If you're using a custom runner or have specific caching needs, you may need to configure cache paths differently.

## Issue: `jq: command not found`

**Symptom**: Workflow fails with "jq: command not found" error

**Solution**: The workflow now includes an `Install jq` step. If you're using a custom runner, ensure jq is installed:

```yaml
- name: Install jq
  run: |
    sudo apt-get update -qq
    sudo apt-get install -qq -y jq
```

## Issue: Git diff fails or shows wrong changes

**Symptom**: `git diff failed` error or incorrect diff analysis

**Possible Causes**:
1. **Shallow clone**: GitHub Actions uses `fetch-depth: 0` to get full history
2. **Wrong refs**: Check that `GITHUB_BASE_SHA` and `GITHUB_HEAD_SHA` are set correctly
3. **Merge-base issues**: The code now uses `git diff base...head` (three dots) for merge-base, falling back to `base..head` if needed

**Solution**: The workflow now handles this automatically. For manual debugging:

```bash
# Check available refs
git log --oneline -10

# Test diff manually
git diff $GITHUB_BASE_SHA...$GITHUB_HEAD_SHA
```

## Issue: Coverage format detection fails

**Symptom**: `failed to detect coverage format` or `unsupported coverage format`

**Solution**: Difftron now supports:
- **LCOV** (`.info` files)
- **Cobertura XML** (`.xml` files)
- **Go coverage** (`.out` files)

The format is auto-detected. If detection fails, check:
1. File exists and is readable
2. File has valid format markers
3. File isn't empty

**Debug**:
```bash
# Check file format
head -5 coverage.out
head -5 coverage.xml

# Test format detection
./bin/difftron analyze --coverage coverage.out
```

## Issue: Coverage file not found

**Symptom**: `Error: coverage.out not found`

**Solution**: Ensure the test step runs before the analysis step and produces coverage:

```yaml
- name: Run tests with coverage
  run: go test -coverprofile=coverage.out ./...

- name: Run difftron CI analysis
  run: ./bin/difftron ci coverage.out
```

## Issue: JSON parsing fails

**Symptom**: `jq` errors or incorrect threshold detection

**Solution**: The workflow now:
1. Installs `jq` automatically
2. Has fallback values if JSON parsing fails
3. Creates fallback JSON if analysis fails

**Debug**:
```bash
# Check JSON structure
cat analysis.json | jq .

# Test JSON parsing
jq -r '.meets_threshold' analysis.json
```

## Issue: No changes detected

**Symptom**: "No changes detected in diff" even when there are changes

**Possible Causes**:
1. **Direct push to main**: For direct pushes, compares against `HEAD~1`
2. **Empty diff**: The changes might be in files not tracked by git
3. **Wrong refs**: Base and head refs might be the same

**Solution**: Check git diff manually:

```bash
# For PRs
git diff $GITHUB_BASE_SHA...$GITHUB_HEAD_SHA

# For direct pushes
git diff HEAD~1 HEAD
```

## Issue: Workflow always passes even with low coverage

**Symptom**: Workflow shows success even when coverage is below threshold

**Possible Causes**:
1. **`continue-on-error: true`**: The difftron step continues even on failure
2. **Exit code not checked**: The gating step might not be running
3. **JSON parsing issue**: `meets_threshold` might not be read correctly

**Solution**: Check the "Gate PRs if threshold not met" step is running and reading the correct value:

```yaml
- name: Gate PRs if threshold not met
  if: always() && github.event_name == 'pull_request' && steps.difftron.outcome != 'skipped'
  run: |
    MEETS_THRESHOLD=$(jq -r '.meets_threshold // false' analysis.json)
    if [ "$MEETS_THRESHOLD" != "true" ]; then
      exit 1
    fi
```

## Issue: Environment variables not set

**Symptom**: Wrong base/head refs detected

**Solution**: The code auto-detects from CI environment:

**GitHub Actions**:
- `GITHUB_BASE_SHA` - Base commit SHA (for PRs)
- `GITHUB_HEAD_SHA` - Head commit SHA (for PRs)
- `GITHUB_SHA` - Current commit SHA
- `GITHUB_BASE_REF` - Base branch name

**GitLab CI**:
- `CI_MERGE_REQUEST_DIFF_BASE_SHA` - Base commit SHA
- `CI_COMMIT_SHA` - Current commit SHA

**Manual override**:
```bash
./bin/difftron ci --base main --head feature-branch coverage.out
```

## Issue: Coverage analysis shows 0% for all files

**Symptom**: All files show 0% coverage even when tests ran

**Possible Causes**:
1. **Path mismatch**: Git diff paths don't match coverage file paths
2. **Coverage file format**: Wrong format or empty coverage
3. **No executable lines**: Files might only have comments/whitespace

**Solution**:
1. Check coverage file has data:
```bash
go tool cover -func=coverage.out | head -10
```

2. Check git diff has changes:
```bash
git diff $BASE...$HEAD --stat
```

3. Verify path matching (coverage uses normalized paths)

## Issue: Workflow times out

**Symptom**: Workflow runs but times out

**Solution**: 
1. Check for stuck processes (like Go test processes)
2. Reduce test timeout:
```yaml
- name: Run tests with coverage
  run: go test -timeout 5m -coverprofile=coverage.out ./...
```

3. Check for infinite loops in tests

## Debugging Tips

### Enable verbose output

Add debug flags to see what's happening:

```yaml
- name: Run difftron CI analysis
  run: |
    ./bin/difftron ci \
      --threshold "$THRESHOLD" \
      --output-file analysis.json \
      coverage.out
    # Debug: Show analysis results
    cat analysis.json | jq .
```

### Check workflow logs

1. Go to Actions tab in GitHub
2. Click on the failed workflow run
3. Expand each step to see logs
4. Look for error messages

### Test locally

Reproduce CI issues locally:

```bash
# Set environment variables like CI
export GITHUB_BASE_SHA=$(git rev-parse HEAD~1)
export GITHUB_HEAD_SHA=$(git rev-parse HEAD)

# Run the same command as CI
./bin/difftron ci --threshold 80 coverage.out
```

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `coverage file not found` | Coverage file doesn't exist | Ensure test step runs first |
| `git diff failed` | Can't generate diff | Check refs are valid, use `fetch-depth: 0` |
| `failed to detect coverage format` | Unrecognized format | Check file format, ensure it's LCOV/Cobertura/Go |
| `failed to parse coverage file` | Invalid coverage data | Verify coverage file is valid |
| `jq: command not found` | jq not installed | Add jq installation step |

## Getting Help

If you encounter an issue not listed here:

1. **Check workflow logs**: Look for specific error messages
2. **Test locally**: Reproduce the issue locally
3. **Check environment**: Verify CI environment variables
4. **Review recent changes**: Check if recent code changes broke CI
5. **Open an issue**: Include workflow logs and error messages
