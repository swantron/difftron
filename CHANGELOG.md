# Changelog

All notable changes to Difftron will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Health subcommand**: Comprehensive testing health analysis with multi-test-type aggregation
  - `difftron health` command with support for unit, API, and functional test coverage
  - Baseline comparison to detect coverage regressions
  - Separate thresholds for new vs modified files
  - PR/MR commenter support (GitHub/GitLab)
  - JSON, Markdown, and structured text output formats
- **Enhanced path normalization**: Cross-platform path handling with `filepath.ToSlash` and repo-root rebasing
- **Line-by-line Go coverage parsing**: Direct parsing of `.out` files (mode: set/count) for accurate coverage
- **Separate thresholds**: `--threshold-new` and `--threshold-modified` flags for different coverage requirements
- **Markdown output**: Added markdown format support to `analyze` command
- **Proper JSON marshaling**: Replaced fmt-based JSON output with structured marshaling via `pkg/report`
- **CI templates**: Reusable GitHub Actions and GitLab CI templates
  - `.github/workflows/difftron-template.yml` - Reusable workflow template
  - `.gitlab-ci/difftron.yml` - GitLab CI template
  - `.github/workflows/dogfood.yml` - Dogfooding workflow
- **Documentation**: Added `HEALTH_COMMAND.md` with comprehensive health command documentation

### Changed
- **go.mod**: Fixed Go version from invalid `1.25.3` to `1.21` (matching CI workflows)
- **Path matching**: Enhanced path matching strategy with multiple fallback attempts
- **JSON output**: Improved JSON structure with new/modified file breakdowns

### Fixed
- Path normalization now properly handles absolute paths and repo-root rebasing
- Go coverage parsing now supports line-by-line ranges instead of function-level only

## [v0.1] - Initial Release

### Added
- Core CLI with `analyze` and `ci` commands
- Git diff parsing (Hunk Engine)
- Multiple coverage format support (LCOV, Cobertura XML, Go coverage)
- Baseline coverage tracking
- Text and JSON output formats
- Coverage threshold checking
- GitLab CI integration
- Comprehensive test coverage
