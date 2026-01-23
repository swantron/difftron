# Test Data

This directory contains test fixtures for difftron.

## Files

- `fixtures/tronswan-coverage.info` - Real LCOV coverage report generated from the tronswan repository
- `fixtures/sample.diff` - Sample git diff for testing

## Usage

These fixtures can be used to test difftron:

```bash
# Test with sample diff
difftron analyze --coverage fixtures/tronswan-coverage.info --diff fixtures/sample.diff

# Test with actual git diff from tronswan repo
cd /path/to/tronswan
git diff HEAD~1 HEAD > /path/to/difftron/testdata/fixtures/tronswan-diff.patch
cd /path/to/difftron
difftron analyze --coverage testdata/fixtures/tronswan-coverage.info --diff testdata/fixtures/tronswan-diff.patch
```

## Regenerating Coverage

To regenerate the coverage file from tronswan:

```bash
cd /path/to/tronswan
yarn test:coverage
cp coverage/lcov.info /path/to/difftron/testdata/fixtures/tronswan-coverage.info
```
