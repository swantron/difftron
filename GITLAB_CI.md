# GitLab CI Integration Guide

This guide covers integrating Difftron into GitLab CI/CD pipelines with security-friendly artifact distribution.

## Overview

Difftron can be integrated into GitLab CI pipelines to:
- Analyze code coverage on merge requests
- Post coverage reports as MR comments
- Gate merges based on coverage thresholds
- Provide holistic testing health analysis

## Security Considerations

For enterprise environments with strict security policies, Difftron supports multiple artifact distribution methods:

### Option 1: Vendor Binary (Recommended for Security)

**Best for**: Organizations requiring full control over binaries, no external downloads

1. **Download and verify binary**:
```bash
# Download from GitHub releases
wget https://github.com/swantron/difftron/releases/download/v0.1.0/difftron-linux-amd64

# Verify checksum
echo "EXPECTED_CHECKSUM  difftron-linux-amd64" | sha256sum -c

# Commit to repository vendor directory
mkdir -p vendor/difftron
cp difftron-linux-amd64 vendor/difftron/difftron
chmod +x vendor/difftron/difftron
git add vendor/difftron/difftron
git commit -m "Add difftron binary to vendor"
```

2. **Use in GitLab CI**:
```yaml
difftron-analysis:
  stage: test
  image: golang:1.21
  script:
    - go test -coverprofile=coverage.out ./...
    - ./vendor/difftron/difftron ci --threshold 80 coverage.out
  only:
    - merge_requests
```

**Security Benefits**:
- ✅ No external network calls during CI
- ✅ Binary is version-controlled and auditable
- ✅ Can be reviewed by security team before committing
- ✅ Works in air-gapped environments

### Option 2: Build from Source

**Best for**: Organizations requiring source code review, no binary dependencies

1. **Add as Git submodule or copy source**:
```bash
# Option A: Git submodule
git submodule add https://github.com/swantron/difftron.git vendor/difftron

# Option B: Copy source (if submodules not allowed)
git clone https://github.com/swantron/difftron.git vendor/difftron
rm -rf vendor/difftron/.git
git add vendor/difftron
git commit -m "Add difftron source to vendor"
```

2. **Build in CI**:
```yaml
difftron-analysis:
  stage: test
  image: golang:1.21
  before_script:
    - cd vendor/difftron
    - go build -o ../bin/difftron ./cmd/difftron
    - cd ../..
  script:
    - go test -coverprofile=coverage.out ./...
    - ./bin/difftron ci --threshold 80 coverage.out
  only:
    - merge_requests
```

**Security Benefits**:
- ✅ Full source code visibility
- ✅ Can audit and modify source code
- ✅ No binary dependencies
- ✅ Build process is transparent

### Option 3: Docker Image (Internal Registry)

**Best for**: Organizations with internal container registries

1. **Build and push to internal registry**:
```bash
# Build Docker image
docker build -t internal-registry.example.com/tools/difftron:v0.1.0 .

# Push to internal registry
docker push internal-registry.example.com/tools/difftron:v0.1.0
```

2. **Use in GitLab CI**:
```yaml
difftron-analysis:
  stage: test
  image: internal-registry.example.com/tools/difftron:v0.1.0
  script:
    - go test -coverprofile=coverage.out ./...
    - difftron ci --threshold 80 coverage.out
  only:
    - merge_requests
```

**Security Benefits**:
- ✅ Uses internal registry (no external access)
- ✅ Image can be scanned and approved by security team
- ✅ Version pinning ensures consistency

### Option 4: Checksum Verification

**Best for**: Organizations allowing external downloads with verification

```yaml
difftron-analysis:
  stage: test
  image: golang:1.21
  before_script:
    # Download and verify checksum
    - |
      DIFTRON_VERSION="v0.1.0"
      DIFTRON_CHECKSUM="abc123def456..." # Get from GitHub releases
      wget https://github.com/swantron/difftron/releases/download/${DIFTRON_VERSION}/difftron-linux-amd64
      echo "${DIFTRON_CHECKSUM}  difftron-linux-amd64" | sha256sum -c
      chmod +x difftron-linux-amd64
      mv difftron-linux-amd64 /usr/local/bin/difftron
  script:
    - go test -coverprofile=coverage.out ./...
    - difftron ci --threshold 80 coverage.out
  only:
    - merge_requests
```

**Security Benefits**:
- ✅ Checksum verification ensures binary integrity
- ✅ Version pinning prevents unexpected updates
- ✅ Can be combined with allowlist of allowed domains

## Complete GitLab CI Configuration

### Basic Coverage Analysis

```yaml
stages:
  - test
  - coverage

variables:
  DIFTRON_THRESHOLD: "80"

# Run tests with coverage
test-with-coverage:
  stage: test
  image: golang:1.21
  script:
    - go test -coverprofile=coverage.out ./...
  artifacts:
    paths:
      - coverage.out
    expire_in: 1 week
  only:
    - merge_requests
    - main

# Analyze coverage with Difftron
difftron-analysis:
  stage: coverage
  image: golang:1.21
  dependencies:
    - test-with-coverage
  before_script:
    # Use vendored binary (Option 1)
    - chmod +x vendor/difftron/difftron || true
    # OR build from source (Option 2)
    # - cd vendor/difftron && go build -o ../bin/difftron ./cmd/difftron && cd ../..
  script:
    - |
      if [ -f vendor/difftron/difftron ]; then
        ./vendor/difftron/difftron ci --threshold ${DIFTRON_THRESHOLD} coverage.out
      elif [ -f bin/difftron ]; then
        ./bin/difftron ci --threshold ${DIFTRON_THRESHOLD} coverage.out
      else
        echo "Error: Difftron binary not found"
        exit 1
      fi
  artifacts:
    paths:
      - analysis.json
    expire_in: 1 week
  only:
    - merge_requests
```

### Holistic Health Analysis

```yaml
stages:
  - test
  - health-analysis

variables:
  DIFTRON_THRESHOLD: "80"

# Unit tests
unit-tests:
  stage: test
  image: golang:1.21
  script:
    - go test -coverprofile=unit-coverage.out ./...
  artifacts:
    paths:
      - unit-coverage.out
    expire_in: 1 week

# API tests
api-tests:
  stage: test
  image: python:3.11
  script:
    - pip install pytest pytest-cov
    - pytest --cov=api --cov-report=lcov api-coverage.info
  artifacts:
    paths:
      - api-coverage.info
    expire_in: 1 week

# Functional tests
functional-tests:
  stage: test
  image: node:18
  script:
    - npm install
    - npm run test:e2e -- --coverage functional-coverage.info
  artifacts:
    paths:
      - functional-coverage.info
    expire_in: 1 week

# Holistic health analysis
difftron-health:
  stage: health-analysis
  image: golang:1.21
  dependencies:
    - unit-tests
    - api-tests
    - functional-tests
  before_script:
    - chmod +x vendor/difftron/difftron || true
  script:
    - |
      ./vendor/difftron/difftron health \
        --unit-coverage unit-coverage.out \
        --api-coverage api-coverage.info \
        --functional-coverage functional-coverage.info \
        --threshold ${DIFTRON_THRESHOLD} \
        --output json > health-report.json \
        --output markdown > health-report.md
  artifacts:
    paths:
      - health-report.json
      - health-report.md
    reports:
      # GitLab will display coverage in MR
      coverage_report:
        coverage_format: cobertura
        path: unit-coverage.out
    expire_in: 1 week
  only:
    - merge_requests
```

### Merge Request Comments

To post coverage reports as MR comments, use GitLab's API:

```yaml
difftron-mr-comment:
  stage: coverage
  image: curlimages/curl:latest
  dependencies:
    - difftron-analysis
  script:
    - |
      # Read analysis JSON
      ANALYSIS=$(cat analysis.json)
      
      # Extract coverage percentage
      COVERAGE=$(echo "$ANALYSIS" | jq -r '.coverage_percentage')
      THRESHOLD=$(echo "$ANALYSIS" | jq -r '.threshold')
      MEETS_THRESHOLD=$(echo "$ANALYSIS" | jq -r '.meets_threshold')
      
      # Create markdown comment
      if [ "$MEETS_THRESHOLD" = "true" ]; then
        EMOJI="✅"
        STATUS="PASS"
      else
        EMOJI="❌"
        STATUS="FAIL"
      fi
      
      COMMENT="## ${EMOJI} Difftron Coverage Analysis
      
      **Coverage:** ${COVERAGE}% (threshold: ${THRESHOLD}%)
      **Status:** ${STATUS}
      
      See [coverage report](${CI_JOB_URL}) for details."
      
      # Post comment to MR
      curl --request POST \
        --header "PRIVATE-TOKEN: ${GITLAB_TOKEN}" \
        --header "Content-Type: application/json" \
        --data "{\"body\": \"${COMMENT}\"}" \
        "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests/${CI_MERGE_REQUEST_IID}/notes"
  only:
    - merge_requests
```

## Artifact Distribution Best Practices

### 1. Version Pinning

Always pin specific versions:
```yaml
variables:
  DIFTRON_VERSION: "v0.1.0"  # Pin to specific version
```

### 2. Checksum Verification

Verify binary integrity:
```bash
# Get checksum from GitHub releases page
DIFTRON_CHECKSUM="abc123def456..."

# Verify before use
echo "${DIFTRON_CHECKSUM}  difftron" | sha256sum -c
```

### 3. Vendor Directory Structure

Recommended vendor structure:
```
vendor/
├── difftron/
│   ├── difftron          # Binary (Option 1)
│   ├── cmd/              # Source (Option 2)
│   ├── internal/
│   └── go.mod
└── checksums.txt         # Checksums for verification
```

### 4. Security Review Checklist

Before adding Difftron to your pipeline:

- [ ] Review source code (if using Option 2)
- [ ] Verify binary checksums (if using Option 1 or 4)
- [ ] Scan Docker image (if using Option 3)
- [ ] Review dependencies (`go.mod` for Go projects)
- [ ] Test in non-production environment first
- [ ] Document approval in security review system
- [ ] Set up monitoring for unexpected behavior

## Troubleshooting

### Binary Not Found

```yaml
# Add check in before_script
before_script:
  - |
    if [ ! -f vendor/difftron/difftron ] && [ ! -f bin/difftron ]; then
      echo "Error: Difftron binary not found"
      echo "Please add difftron to vendor directory or build from source"
      exit 1
    fi
```

### Permission Denied

```yaml
before_script:
  - chmod +x vendor/difftron/difftron
```

### Coverage File Not Found

```yaml
script:
  - |
    if [ ! -f coverage.out ]; then
      echo "Error: coverage.out not found"
      echo "Make sure test job runs before difftron job"
      exit 1
    fi
```

## Additional Resources

- [Difftron Documentation](README.md)
- [Holistic Health Reporting](HOLISTIC_HEALTH.md)
- [Baseline Coverage Tracking](BASELINE_COVERAGE.md)
- [GitLab CI/CD Documentation](https://docs.gitlab.com/ee/ci/)
