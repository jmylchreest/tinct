# Tinct Project Setup Guide

This document provides an overview of the project configuration and setup for the Tinct colour palette generator.

## Overview

Tinct is built with **Go 1.25+** and uses modern Go features including generics, structured logging (slog), range over functions, and enhanced error handling. The project follows SOLID principles and idiomatic Go practices.

## Project Structure

```
tinct/
├── .github/
│   └── workflows/
│       └── release.yml          # CI/CD with GitHub Actions
├── cmd/
│   └── tinct/                   # Main application entry point
├── internal/                    # Private application code
│   ├── app/                     # Application orchestration
│   ├── colour/                   # Colour extraction & palette generation
│   ├── config/                  # Configuration management
│   ├── plugin/                  # Plugin system (Source & Output)
│   ├── hook/                    # Lifecycle hooks
│   └── image/                   # Image loading & processing
├── pkg/                         # Public API
├── configs/                     # Example configurations
├── docs/                        # Additional documentation
├── completions/                 # Shell completions (generated)
├── .goreleaser.yml             # GoReleaser configuration
├── .golangci.yml               # Linter configuration
├── Dockerfile                   # Multi-stage container build
├── docs/
│   ├── DEVELOPMENT.md          # Development guidelines (LLM instructions)
├── README.md                    # User documentation
│   └── SETUP.md                 # This file
```

## Configuration Files

### .goreleaser.yml
Automated release configuration supporting:
- **Multi-platform builds**: Linux, macOS, Windows, FreeBSD (amd64, arm64, arm)
- **Package managers**: Homebrew, AUR (Arch), Nix, APT (Debian), RPM, APK
- **Signing**: Cosign for artifact signing
- **SBOMs**: Software Bill of Materials generation
- **Checksums**: SHA256 checksums for all artifacts

**Environment variables needed for full release:**
```bash
GITHUB_TOKEN              # GitHub releases (automatic in Actions)
HOMEBREW_TAP_GITHUB_TOKEN # For Homebrew tap updates
AUR_KEY                   # For AUR package updates
NIX_GITHUB_TOKEN          # For Nix package updates
COSIGN_EXPERIMENTAL=1     # For keyless signing
```

### .golangci.yml
Comprehensive linting with 30+ linters enabled:
- **Correctness**: errcheck, govet, staticcheck, typecheck
- **Style**: gofmt, goimports, revive, gocritic
- **Performance**: prealloc
- **Security**: gosec
- **Complexity**: gocyclo, gocognit, nestif
- **Best practices**: errorlint, wrapcheck, nolintlint

Configured for Go 1.25 with custom rules for test files.

### GitHub Actions (.github/workflows/release.yml)
CI/CD pipeline with multiple jobs:
- **Test**: Unit tests, race detection, coverage (Codecov)
- **Lint**: golangci-lint, go vet, formatting checks
- **Security**: Gosec, Trivy vulnerability scanning
- **Build**: Snapshot builds for all platforms
- **Release**: Full release with GoReleaser (tag-triggered)

## Getting Started

### Prerequisites

1. **Go 1.25+** - [Download](https://go.dev/dl/)
   ```bash
   go version  # Should show 1.25 or higher
   ```

2. **Development Tools** (optional but recommended):
   ```bash
   go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
   go install github.com/goreleaser/goreleaser@latest
   ```

### Initial Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/jmylchreest/tinct.git
   cd tinct
   ```

2. **Install dependencies**:
   ```bash
   go mod download && go mod verify
   ```

3. **Verify setup**:
   ```bash
   go fmt ./...
   go vet ./...
   golangci-lint run
   go test ./...
   ```

4. **Build**:
   ```bash
   go build -o tinct ./cmd/tinct
   # Binary: ./tinct
   ```

5. **Run**:
   ```bash
   ./tinct --help
   ```

## Development Workflow

### Day-to-Day Development

```bash
# Format code
go fmt ./...
goimports -w .

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run linters
golangci-lint run

# Build and run
go build -o tinct ./cmd/tinct
./tinct generate image.jpg
```

### Building Docker Images

```bash
# Build Docker image
docker build -t tinct:latest .

# Run in container
docker run --rm -it tinct:latest --help
```

### Testing Releases Locally

```bash
# Create snapshot release (no tag needed)
goreleaser release --snapshot --clean

# Check dist/ directory for artifacts
ls -la dist/

# Test release process (dry run)
goreleaser release --snapshot --clean --skip=publish
```

### Creating a Release

```bash
# 1. Ensure working directory is clean
git status

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. GitHub Actions automatically:
#    - Runs tests and lints
#    - Builds for all platforms
#    - Creates GitHub release
#    - Updates package managers (if configured)
```

## Plugin System Architecture

Tinct uses a **dual plugin system**:

### 1. Source Plugins
Generate or fetch wallpapers:
- **AI Wallpaper**: Generate from text prompts (DALL-E, Stable Diffusion)
- **Random Fetch**: Download from Unsplash, Pexels, etc.
- **Local Random**: Select from local directory
- **Custom**: Implement `SourcePlugin` interface

### 2. Output Plugins
Generate application config files:
- **Terminal Emulators**: Alacritty, Kitty
- **Window Managers**: i3, Sway
- **Application Launchers**: Rofi
- **Status Bars**: Polybar
- **Notification Daemons**: Dunst
- **Custom**: Implement `OutputPlugin` interface

**Plugin Flow**:
```
Source Plugin → Image → Colour Extraction → Palette → Output Plugins
     ↓                                                      ↓
  (Optional)                                    (Configuration Files)
```

## Hook System

Hooks provide lifecycle management:

- **Pre-Execution**: Readiness checks, API validation, backups
- **Post-Execution**: Service reloads, wallpaper setting, notifications
- **On Error**: Rollback, cleanup, logging

See [DEVELOPMENT.md](DEVELOPMENT.md) for implementation details.

## Documentation

- **[Development Guide](DEVELOPMENT.md)**: Comprehensive development guidelines
  - Go 1.25+ features and best practices
  - SOLID principles in Go
  - Plugin development guide
  - Hook system implementation
  - Testing strategies
  - Error handling patterns

- **[README.md](README.md)**: User-facing documentation
  - Installation instructions
  - Usage examples
  - Configuration guide
  - Plugin list and custom plugin creation
  - Architecture overview

## Go 1.25+ Features Used

Tinct leverages modern Go capabilities:

1. **Generics**: Type-safe plugin registry
2. **Range Over Functions**: Custom iterators for palette colours
3. **Structured Logging (log/slog)**: JSON/text logging with levels
4. **Enhanced Errors**: `errors.Join()` for multiple errors
5. **Slices Package**: `slices.Clone()`, `slices.Sort()`, etc.
6. **Maps Package**: `maps.Clone()`, `maps.Equal()`, etc.
7. **Cmp Package**: Generic comparison functions

**Reference**: [Go 1.25 Language Specification](https://go.dev/ref/spec)

## Environment Variables

### Development
```bash
# AI API Configuration
export TINCT_AI_API_KEY=your_api_key
export TINCT_AI_PROVIDER=stablediffusion  # dalle, stablediffusion

# Configuration
export TINCT_CONFIG=~/.config/tinct/config.yaml
export TINCT_ALGORITHM=kmeans
export TINCT_COLORS=16
export TINCT_OUTPUTS=alacritty,kitty

# Development
export TINCT_LOG_LEVEL=debug
```

### CI/CD (GitHub Secrets)
```bash
GITHUB_TOKEN              # Automatic in GitHub Actions
HOMEBREW_TAP_GITHUB_TOKEN # Optional: Homebrew tap
AUR_KEY                   # Optional: AUR packages
NIX_GITHUB_TOKEN          # Optional: Nix packages
CODECOV_TOKEN             # Optional: Code coverage
```

## Docker Usage

### Build
```bash
docker build -t tinct:latest .
```

### Run
```bash
# Interactive
docker run --rm -it tinct:latest --help

# With volume mounts
docker run --rm -v $PWD/images:/images -v $PWD/output:/output \
  tinct:latest generate /images/wallpaper.jpg

# From GitHub Container Registry
docker pull ghcr.io/jmylchreest/tinct:latest
```

## Package Installation (Post-Release)

### Homebrew (macOS/Linux)
```bash
brew install jmylchreest/tap/tinct
```

### AUR (Arch Linux)
```bash
yay -S tinct-bin
# or
paru -S tinct-bin
```

### APT (Debian/Ubuntu)
```bash
# Download .deb from releases
sudo dpkg -i tinct_*.deb
```

### RPM (Fedora/RHEL)
```bash
# Download .rpm from releases
sudo rpm -i tinct_*.rpm
```

### Go Install
```bash
go install github.com/jmylchreest/tinct/cmd/tinct@latest
```

## Testing Strategy

- **Unit Tests**: Test individual components
- **Integration Tests**: Test plugin system and hooks
- **Table-Driven Tests**: Comprehensive input/output coverage
- **Benchmarks**: Performance testing for colour extraction
- **Race Detection**: Concurrent execution safety
- **Coverage Target**: >80% for critical paths, 100% for public APIs

Run tests:
```bash
go test ./...                                    # Basic tests
go test -cover ./...                            # With coverage
go test -coverprofile=coverage.out ./...       # Generate coverage file
go tool cover -html=coverage.out                # HTML report
go test -bench=. -benchmem ./...               # Benchmarks
```

## Security

- **Gosec**: Static security analysis
- **Trivy**: Dependency vulnerability scanning
- **Cosign**: Artifact signing (keyless)
- **SBOM**: Software Bill of Materials generation
- **Dependabot**: Automated dependency updates (configure in GitHub)

## Commit Convention

Use conventional commits for automatic changelog generation:

- `feat:` - New features
- `fix:` - Bug fixes
- `perf:` - Performance improvements
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `chore:` - Maintenance tasks
- `ci:` - CI/CD changes

Example:
```bash
git commit -m "feat(plugin): add AI wallpaper generation support"
git commit -m "fix(extract): handle edge case in k-means clustering"
```

## Contributing

1. Read [DEVELOPMENT.md](DEVELOPMENT.md)
2. Fork the repository
3. Create feature branch: `git checkout -b feat/amazing-feature`
4. Make changes following guidelines
5. Run tests and linting: `go test ./... && golangci-lint run`
6. Commit using conventional commits
7. Push and create pull request

## Support

- **Issues**: [GitHub Issues](https://github.com/jmylchreest/tinct/issues)
- **Discussions**: [GitHub Discussions](https://github.com/jmylchreest/tinct/discussions)
- **Email**: John Mylchreest <jmylchreest@gmail.com>

## License

MIT License - See [LICENSE](LICENSE) file

---

**Happy theming!**