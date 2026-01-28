# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Security Considerations for Enterprise Use

Difftron is designed to be security-friendly for enterprise environments. This document outlines security considerations and best practices.

### What Difftron Does

Difftron is a **read-only analysis tool** that:
- Reads git diff output (from `git diff` command)
- Reads coverage files (LCOV, Go coverage format)
- Analyzes coverage data locally
- Outputs reports (JSON, Markdown, text)
- **Does NOT** make external network calls during analysis
- **Does NOT** modify your code or repository
- **Does NOT** execute your tests (only reads coverage reports)

### Network Access

Difftron does **not** require network access during normal operation:
- ✅ All analysis is performed locally
- ✅ No external API calls during coverage analysis
- ✅ No telemetry or tracking
- ✅ No data sent to external services

**Note**: Future AI features (planned) may require API access, but will be opt-in and clearly documented.

### Artifact Distribution

For security-conscious environments, we recommend:

1. **Vendor the binary** (see [GITLAB_CI.md](GITLAB_CI.md))
   - Download from GitHub releases
   - Verify checksums
   - Commit to your repository's vendor directory
   - No external downloads during CI

2. **Build from source** (see [GITLAB_CI.md](GITLAB_CI.md))
   - Copy source code to vendor directory
   - Build in CI pipeline
   - Full source code visibility

3. **Use internal Docker registry**
   - Build Docker image
   - Push to internal registry
   - Use in CI pipelines

### Checksum Verification

All releases include SHA256 checksums:

```bash
# Verify binary integrity
wget https://github.com/swantron/difftron/releases/download/v0.1.0/difftron-linux-amd64
wget https://github.com/swantron/difftron/releases/download/v0.1.0/difftron-linux-amd64.sha256
sha256sum -c difftron-linux-amd64.sha256
```

### Dependencies

Difftron has minimal dependencies:
- Standard Go library (no external packages for core functionality)
- Cobra CLI framework (for CLI interface)
- All dependencies are pinned in `go.sum`

Review dependencies:
```bash
go list -m all
```

### Code Review

The entire codebase is open source and can be reviewed:
- Source code: `internal/` directory
- Tests: `*_test.go` files
- No obfuscated or minified code
- No hidden functionality

### Permissions

Difftron requires minimal permissions:
- **Read access**: Coverage files, git repository
- **Write access**: Output files (optional, can write to stdout)
- **No special privileges**: Runs as regular user
- **No file system access**: Only reads specified files

### Reporting Security Issues

If you discover a security vulnerability, please:

1. **Do NOT** open a public issue
2. Email security concerns to: [security contact]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within 48 hours and work with you to resolve the issue.

### Security Best Practices

When using Difftron in enterprise environments:

1. **Vendor artifacts**: Don't download binaries during CI
2. **Verify checksums**: Always verify binary integrity
3. **Pin versions**: Use specific version tags, not `latest`
4. **Review source**: Review source code before use
5. **Scan dependencies**: Scan `go.mod` for known vulnerabilities
6. **Limit permissions**: Run with minimal required permissions
7. **Monitor usage**: Monitor for unexpected behavior
8. **Keep updated**: Update to latest versions with security fixes

### Compliance

Difftron:
- ✅ Does not collect or store user data
- ✅ Does not make external network calls
- ✅ Can run in air-gapped environments
- ✅ Source code is auditable
- ✅ No telemetry or tracking

### Questions?

For security-related questions:
- Review [GITLAB_CI.md](GITLAB_CI.md) for secure integration options
- Check [README.md](README.md) for usage documentation
- Open an issue for general questions (not security vulnerabilities)
