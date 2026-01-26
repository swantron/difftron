# Cache Warnings Explanation

## "Cannot open: File exists" Warnings

If you see multiple "Cannot open: File exists" warnings in your GitHub Actions workflow logs, **these are safe to ignore**.

### What causes these warnings?

These warnings occur when the GitHub Actions cache action (`actions/cache@v4`) tries to restore cached files/directories, but some paths already exist in the filesystem. This can happen when:

1. The cache action runs multiple times in the same workflow
2. Directories are created by other steps before cache restore
3. The runner has leftover files from previous runs

### Are they a problem?

**No.** These are warnings, not errors. The workflow will:
- ✅ Still succeed (workflow shows "Success")
- ✅ Cache will still work (restore happens despite warnings)
- ✅ Build and tests will run normally

### What we've done

The workflow includes `continue-on-error: true` on the cache step to ensure these warnings don't fail the workflow. We also ensure directories exist before `go mod download` to prevent conflicts.

### Should you fix them?

**No action needed** if:
- Workflow status is "Success"
- Tests and builds complete successfully
- Only seeing warnings (not errors)

**Consider investigating** if:
- Workflow is actually failing
- Cache isn't working (builds are slow)
- You see actual errors (not just warnings)

### Example

```
Run actions/cache@v4
Cannot open: File exists
Cannot open: File exists
...
Post job cleanup.
```

This is **normal** and the workflow will still succeed.
