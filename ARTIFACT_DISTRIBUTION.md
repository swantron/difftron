# Artifact Distribution Guide

Quick reference for security-friendly artifact distribution options for Difftron.

## Quick Decision Tree

```
Do you allow external downloads?
├─ NO → Use Option 1 (Vendor Binary) or Option 2 (Build from Source)
└─ YES → Use Option 4 (Checksum Verification)

Do you have an internal container registry?
└─ YES → Use Option 3 (Docker Image)

Do you need source code visibility?
└─ YES → Use Option 2 (Build from Source)
```

## Option Comparison

| Option | External Network | Source Visible | Air-Gapped | Security Review |
|--------|-----------------|---------------|------------|-----------------|
| 1. Vendor Binary | ❌ No | ❌ No | ✅ Yes | Binary only |
| 2. Build from Source | ❌ No | ✅ Yes | ✅ Yes | Full source |
| 3. Docker Image | ❌ No* | ⚠️ Optional | ✅ Yes | Image scan |
| 4. Checksum Verify | ✅ Yes | ❌ No | ❌ No | Binary + checksum |

*Uses internal registry, not external

## Option 1: Vendor Binary (Recommended)

**Best for**: Most enterprise environments

### Setup Steps

1. **Download and verify**:
```bash
# Download from GitHub releases
VERSION="v0.1.0"
wget https://github.com/swantron/difftron/releases/download/${VERSION}/difftron-linux-amd64
wget https://github.com/swantron/difftron/releases/download/${VERSION}/difftron-linux-amd64.sha256

# Verify checksum
sha256sum -c difftron-linux-amd64.sha256

# Add to vendor directory
mkdir -p vendor/difftron
cp difftron-linux-amd64 vendor/difftron/difftron
chmod +x vendor/difftron/difftron

# Commit to repository
git add vendor/difftron/difftron
git commit -m "Add difftron v${VERSION} to vendor"
```

2. **Use in CI**:
```yaml
script:
  - ./vendor/difftron/difftron ci coverage.out
```

### Pros
- ✅ No external network calls during CI
- ✅ Version-controlled and auditable
- ✅ Works in air-gapped environments
- ✅ Fast execution (no build time)

### Cons
- ❌ Binary not human-readable
- ❌ Requires manual updates

## Option 2: Build from Source

**Best for**: Organizations requiring source code review

### Setup Steps

1. **Add source to vendor**:
```bash
# Clone source
git clone https://github.com/swantron/difftron.git vendor/difftron
cd vendor/difftron
git checkout v0.1.0  # Pin to specific version
cd ../..

# Remove .git to avoid submodule complexity
rm -rf vendor/difftron/.git

# Commit to repository
git add vendor/difftron
git commit -m "Add difftron source v0.1.0 to vendor"
```

2. **Build in CI**:
```yaml
before_script:
  - cd vendor/difftron
  - go build -o ../../bin/difftron ./cmd/difftron
  - cd ../..
script:
  - ./bin/difftron ci coverage.out
```

### Pros
- ✅ Full source code visibility
- ✅ Can audit and modify source
- ✅ No binary dependencies
- ✅ Build process is transparent

### Cons
- ❌ Requires Go toolchain in CI
- ❌ Slower (build time)

## Option 3: Docker Image (Internal Registry)

**Best for**: Organizations with internal container registries

### Setup Steps

1. **Build and push to internal registry**:
```bash
# Build image
docker build -t internal-registry.example.com/tools/difftron:v0.1.0 .

# Scan image (if required)
docker scan internal-registry.example.com/tools/difftron:v0.1.0

# Push to internal registry
docker push internal-registry.example.com/tools/difftron:v0.1.0
```

2. **Use in CI**:
```yaml
difftron-analysis:
  image: internal-registry.example.com/tools/difftron:v0.1.0
  script:
    - difftron ci coverage.out
```

### Pros
- ✅ Uses internal registry
- ✅ Image can be scanned and approved
- ✅ Version pinning
- ✅ Consistent environment

### Cons
- ❌ Requires Docker registry setup
- ❌ Requires image scanning process

## Option 4: Checksum Verification

**Best for**: Organizations allowing external downloads with verification

### Setup Steps

1. **Get checksum from releases**:
```bash
# Download checksum file
wget https://github.com/swantron/difftron/releases/download/v0.1.0/difftron-linux-amd64.sha256
cat difftron-linux-amd64.sha256
# Output: abc123def456...  difftron-linux-amd64
```

2. **Use in CI**:
```yaml
before_script:
  - |
    VERSION="v0.1.0"
    CHECKSUM="abc123def456..."  # From releases page
    wget https://github.com/swantron/difftron/releases/download/${VERSION}/difftron-linux-amd64
    echo "${CHECKSUM}  difftron-linux-amd64" | sha256sum -c
    chmod +x difftron-linux-amd64
    mv difftron-linux-amd64 /usr/local/bin/difftron
script:
  - difftron ci coverage.out
```

### Pros
- ✅ Automatic updates (change version)
- ✅ Checksum verification
- ✅ No manual vendor management

### Cons
- ❌ Requires external network access
- ❌ Not suitable for air-gapped environments

## Security Review Checklist

Before adding Difftron to your pipeline:

- [ ] Choose distribution option based on security requirements
- [ ] Review source code (if using Option 2)
- [ ] Verify binary checksums (if using Option 1 or 4)
- [ ] Scan Docker image (if using Option 3)
- [ ] Review dependencies (`go.mod`)
- [ ] Test in non-production environment
- [ ] Document approval in security review system
- [ ] Set up monitoring for unexpected behavior
- [ ] Pin to specific version (not `latest`)
- [ ] Document update process

## Version Updates

### For Option 1 (Vendor Binary)

```bash
# Update process
VERSION="v0.2.0"
wget https://github.com/swantron/difftron/releases/download/${VERSION}/difftron-linux-amd64
wget https://github.com/swantron/difftron/releases/download/${VERSION}/difftron-linux-amd64.sha256
sha256sum -c difftron-linux-amd64.sha256
cp difftron-linux-amd64 vendor/difftron/difftron
git add vendor/difftron/difftron
git commit -m "Update difftron to ${VERSION}"
```

### For Option 2 (Build from Source)

```bash
# Update process
cd vendor/difftron
git fetch origin
git checkout v0.2.0
cd ../..
git add vendor/difftron
git commit -m "Update difftron to v0.2.0"
```

### For Option 3 (Docker Image)

```bash
# Update process
VERSION="v0.2.0"
docker build -t internal-registry.example.com/tools/difftron:${VERSION} .
docker scan internal-registry.example.com/tools/difftron:${VERSION}
docker push internal-registry.example.com/tools/difftron:${VERSION}
# Update CI to use new version
```

## Troubleshooting

### Binary Not Found

```bash
# Check if binary exists
ls -la vendor/difftron/difftron

# Check permissions
chmod +x vendor/difftron/difftron

# Verify path in CI
pwd
ls -la
```

### Checksum Mismatch

```bash
# Re-download and verify
rm difftron-linux-amd64
wget https://github.com/swantron/difftron/releases/download/v0.1.0/difftron-linux-amd64
sha256sum difftron-linux-amd64
# Compare with expected checksum from releases page
```

### Build Failures (Option 2)

```bash
# Check Go version
go version  # Requires Go 1.21+

# Check dependencies
cd vendor/difftron
go mod download
go build ./cmd/difftron
```

## Additional Resources

- [GITLAB_CI.md](GITLAB_CI.md): Complete GitLab CI integration guide
- [SECURITY.md](SECURITY.md): Security policy and best practices
- [GitHub Releases](https://github.com/swantron/difftron/releases): Download binaries and checksums
