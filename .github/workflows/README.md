# GitHub Actions Workflows

## Difftron Quality Gate

The `difftron.yml` workflow automatically analyzes code coverage on pull requests and pushes to main, **gating** the workflow if coverage threshold is not met.

### Features

- **PR Comments**: Automatically posts coverage analysis as a comment on pull requests
- **Status Checks**: Creates check runs that show up in the PR checks section
- **Auto-Update**: Comments are updated on each push to the PR (no duplicate comments)
- **Gating**: **Fails the workflow if coverage threshold is not met** (blocks merges/pushes)
- **Configurable Threshold**: Set threshold via workflow input, PR labels, or default (80%)

### Configuration

#### Default Threshold

The default threshold is **80%**. This can be changed in the workflow file:

```yaml
- name: Set threshold
  run: |
    THRESHOLD="${INPUT_THRESHOLD:-80}"  # Change 80 to your desired default
```

#### Per-Run Threshold (Manual Trigger)

When manually triggering the workflow, you can set a custom threshold:

1. Go to Actions → Difftron Quality Gate → Run workflow
2. Enter your threshold percentage (e.g., `90`)

#### Per-PR Threshold (PR Labels)

Add a label to your PR with format: `coverage-threshold:90`

Example labels:
- `coverage-threshold:90` - Sets threshold to 90%
- `coverage-threshold:100` - Requires 100% coverage

### How It Works

1. **On Pull Request**:
   - Triggers on: `opened`, `synchronize`, `reopened`
   - Compares PR branch against base branch (main)
   - Posts comment with coverage analysis
   - Creates check run for PR status
   - **Gates the PR** - PR cannot be merged if threshold not met

2. **On Push to Main**:
   - Compares current commit against previous commit
   - **Fails the workflow** if threshold not met (blocks the push/merge)

3. **Manual Trigger**:
   - Can be run manually with custom threshold
   - Useful for testing or one-off analysis

### Gating Behavior

**For Pull Requests:**
- **Will fail** (gate/block merge) if coverage is below threshold
- **Will pass** if coverage meets threshold or no changes detected

**For Direct Pushes to Main:**
- **Always passes** (informational only)
- Reports coverage status but does not block
- Useful for monitoring coverage trends without blocking deployments

**Rationale**: PRs should be gated to prevent merging untested code. Direct pushes to main are typically from trusted sources (like your comprehensive pipelines), so gating would be too restrictive.

### Example PR Comment

```
## ✅ Difftron Coverage Analysis

**Coverage:** 85.7% (threshold: 80%)  
**Status:** PASS

### Summary
- **Total Changed Lines:** 12
- **Covered Lines:** 10
- **Uncovered Lines:** 2

### Files
- **internal/analyzer/analyzer.go**: 100.0% coverage
- **cmd/difftron/ci.go**: 75.0% coverage (uncovered lines: 42, 45)
```

### Customization

Edit `.github/workflows/difftron.yml` to customize:

- **Default Threshold**: Change the default value in the "Set threshold" step
- **Branches**: Modify `branches: [main]` to run on other branches
- **Events**: Add/remove event types in the `on:` section
- **Comment Format**: Modify the comment body in the "Comment on Pull Request" step
- **Gating**: The "Gate: Fail if threshold not met" step controls gating behavior

### Troubleshooting

**PR not gating when it should?**
- Verify you're testing a PR (not a direct push)
- Check that the "Gate: Fail if threshold not met (PRs only)" step is running
- Verify `meets_threshold` in analysis.json is `false`
- Check workflow logs for the actual coverage percentage

**Direct push not reporting?**
- Check the "Report coverage status (direct pushes)" step logs
- Verify analysis.json was created
- Note: Direct pushes never gate - they're informational only

**Comment not appearing?**
- Check that the workflow has `pull-requests: write` permission
- Verify the PR is targeting the `main` branch
- Check workflow logs for errors

**Threshold not being respected?**
- Verify the threshold is being passed correctly to `difftron ci`
- Check the "Set threshold" step output in workflow logs
- Ensure PR labels follow the format `coverage-threshold:XX`

### Disabling Gating (Not Recommended)

If you want to disable gating (workflow always passes), remove or comment out the "Gate: Fail if threshold not met" step. However, this defeats the purpose of a quality gate.
