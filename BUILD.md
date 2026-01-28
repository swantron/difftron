# Build and Artifact Distribution

This document describes how to build Difftron artifacts and distribute them for use in other build systems.

## Building Artifacts

### Local Build

Build for your current platform:

```bash
go build -o bin/difftron ./cmd/difftron
```

### Cross-Platform Builds

Build for specific platforms:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o difftron-linux-amd64 ./cmd/difftron

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o difftron-linux-arm64 ./cmd/difftron

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o difftron-darwin-amd64 ./cmd/difftron

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o difftron-darwin-arm64 ./cmd/difftron

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o difftron-windows-amd64.exe ./cmd/difftron
```

### Release Builds

The `.github/workflows/release.yml` workflow automatically builds artifacts for all platforms when you push a tag starting with `v`:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The workflow will:
- Build binaries for Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64)
- Create a GitHub Release with all artifacts
- Generate release notes automatically

## Artifact Structure

Each release includes:

```
difftron-linux-amd64
difftron-linux-arm64
difftron-darwin-amd64
difftron-darwin-arm64
difftron-windows-amd64.exe
```

## Using Artifacts in Other Build Systems

### Option 1: Download from GitHub Releases

```bash
VERSION="v1.0.0"
PLATFORM="linux-amd64"

# Download binary
wget https://github.com/swantron/difftron/releases/download/${VERSION}/difftron-${PLATFORM}

# Make executable
chmod +x difftron-${PLATFORM}

# Use it
./difftron-${PLATFORM} ci --threshold 80 coverage.out
```

### Option 2: Vendor Binary in Repository

For enterprise environments that don't allow external downloads:

```bash
# Download binary (see Option 1)
# Then commit to your repository
mkdir -p vendor/difftron
cp difftron-linux-amd64 vendor/difftron/difftron
chmod +x vendor/difftron/difftron
git add vendor/difftron/difftron
git commit -m "Add difftron v1.0.0 to vendor"

# Use in CI/CD
./vendor/difftron/difftron ci --threshold 80 coverage.out
```

### Option 3: Build from Source

For maximum security and source visibility:

```bash
# Clone repository
git clone https://github.com/swantron/difftron.git vendor/difftron
cd vendor/difftron
git checkout v1.0.0

# Build
go build -o difftron ./cmd/difftron

# Use
./difftron ci --threshold 80 coverage.out
```

### Option 4: Docker Image

Build a Docker image:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o difftron ./cmd/difftron

FROM alpine:latest
COPY --from=builder /build/difftron /usr/local/bin/difftron
ENTRYPOINT ["difftron"]
```

## Version Information

Binaries include version information:

```bash
./difftron --version
# Output: dev (commit: unknown, date: unknown)
```

Note: Version information is set at build time. For release builds, the version will reflect the git tag.

### Embedding Version Information

Version information can be embedded at build time using ldflags:

```bash
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
  -o difftron ./cmd/difftron
```

The release workflow automatically embeds version information when building from tags.

## CI/CD Integration

### GitHub Actions

```yaml
- name: Download difftron
  uses: actions/download-artifact@v4
  with:
    name: difftron-linux-amd64
    path: bin/

- name: Run difftron
  run: ./bin/difftron ci --threshold 80 coverage.out
```

### GitLab CI

```yaml
difftron:
  script:
    - ./vendor/difftron/difftron ci --threshold 80 coverage.out
```

See `ARTIFACT_DISTRIBUTION.md` and `GITLAB_CI.md` for detailed integration guides.

## Security Considerations

- ✅ Binaries are statically linked for portability
- ✅ No external dependencies required at runtime
- ✅ Source code available for security review
- ✅ Simple, reproducible builds

## Troubleshooting

**Issue**: Binary won't execute
- **Solution**: Ensure it's executable: `chmod +x difftron-linux-amd64`

**Issue**: Wrong architecture
- **Solution**: Download the correct binary for your platform (check `uname -m`)

**Issue**: Version shows "dev"
- **Solution**: This is expected for local builds. Release builds will show the git tag version
