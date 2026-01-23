# GitHub Actions Workflows

## Difftron Quality Gate

The `difftron.yml` workflow automatically analyzes code coverage on pull requests and pushes to main.

### Features

- **PR Comments**: Automatically posts coverage analysis as a comment on pull requests
- **Status Checks**: Creates check runs that show up in the PR checks section
- **Auto-Update**: Comments are updated on each push to the PR (no duplicate comments)
- **Artifacts**: Uploads coverage reports and analysis results

### How It Works

1. **On Pull Request**:
   - Triggers on: `opened`, `synchronize`, `reopened`
   - Compares PR branch against base branch (main)
   - Posts comment with coverage analysis
   - Creates check run for PR status

2. **On Push to Main**:
   - Compares current commit against previous commit
   - Sets exit code based on coverage threshold
   - Fails the workflow if threshold not met

### Permissions

The workflow requires these permissions:
- `contents: read` - Read repository contents
- `pull-requests: write` - Post PR comments
- `checks: write` - Create check runs

These are automatically granted via `GITHUB_TOKEN` - no additional setup needed!

### Customization

Edit `.github/workflows/difftron.yml` to customize:

- **Threshold**: Change `--threshold 80` to your desired percentage
- **Branches**: Modify `branches: [main]` to run on other branches
- **Events**: Add/remove event types in the `on:` section
- **Comment Format**: Modify the comment body in the "Comment on Pull Request" step

### Example PR Comment

```
## ✅ Difftron Coverage Analysis

**Coverage:** 85.7% (threshold: 80%)  
**Status:** PASS

### Summary
- **Total Changed Lines:** 12
- **Covered Lines:** 10
- **Uncovered Lines:** 2

<details>
<summary>📊 Detailed Report</summary>

[Full analysis output]

</details>
```

### Troubleshooting

**Comment not appearing?**
- Check that the workflow has `pull-requests: write` permission
- Verify the PR is targeting the `main` branch
- Check workflow logs for errors

**Analysis fails?**
- Ensure tests are passing (`go test ./...`)
- Check that coverage.out is generated
- Verify git diff is not empty

**Check run not showing?**
- Verify `checks: write` permission is set
- Check that the PR head SHA is correct
- Review workflow logs for API errors
