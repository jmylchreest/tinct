# Tinct Project Summary

**Version:** 1.0.0  
**Go Version:** 1.25+  
**Author:** John Mylchreest <jmylchreest@gmail.com>  
**License:** MIT  

---

##  Project Overview

**Tinct** is a modern, extensible CLI tool written in Go 1.25+ that generates colour palettes and applies them system-wide. It features:

- **AI-Powered Wallpaper Generation**: Create wallpapers from text prompts using DALL-E, Stable Diffusion, etc.
- **Custom Wallpaper Sizes**: Support for 4K (default), ultrawide, multi-monitor (dual, triple), and custom dimensions
- **Dual Plugin System**: Source plugins (wallpaper generation/fetching) and Output plugins (config file generation)
- **External Plugin Support**: Load third-party plugins at runtime (Go, Python, Shell, Rust, etc.)
- **Multiple Color Extraction Algorithms**: K-means, median cut, dominant color
- **Lifecycle Hooks**: Pre/post execution for readiness checks, reloads, and cleanup
- **Modern Go Architecture**: SOLID principles, generics, structured logging, comprehensive testing

##  Project Structure

```
tinct/
 .github/workflows/
    release.yml              # CI/CD pipeline with GitHub Actions
 cmd/tinct/                   # Main application entry point
 internal/                    # Private application code
    app/                     # Application orchestration
    color/                   # Colour extraction & palette generation
    config/                  # Configuration management (Viper)
    plugin/                  # Dual plugin system
       source/              # Source plugins (AI, fetch, local)
       output/              # Output plugins (alacritty, kitty, etc.)
    hook/                    # Lifecycle hooks
    image/                   # Image loading & processing
 pkg/                         # Public API
 configs/                     # Example configurations
 docs/                        # Documentation
    ai-wallpaper.md         # AI wallpaper generation guide
 completions/                 # Shell completions (generated)
 .goreleaser.yml             # Automated release configuration
 .golangci.yml               # Comprehensive linting config
 Dockerfile                   # Multi-stage container build
 DEVELOPMENT.md              # LLM & developer guidelines
 README.md                    # User documentation
 SETUP.md                     # Setup and configuration guide
 PROJECT_SUMMARY.md          # This file
```

##  Key Features

### 1. AI Wallpaper Generation
- Generate wallpapers from text prompts
- Support for multiple AI providers (Stable Diffusion, DALL-E, Midjourney)
- **Flexible sizing**: 4K (default), ultrawide (21:9), superwide (32:9), dual/triple monitor, custom dimensions
- Automatic palette extraction from generated images
- Caching to reduce API costs
- Cost optimization strategies

### 2. Dual Plugin System

**Built-in Source Plugins** (Wallpaper Input):
- `ai-wallpaper`: Generate with AI from text prompts
- `random-fetch`: Fetch from Unsplash, Pexels, etc.
- `local-random`: Randomly select from local directory

**Built-in Output Plugins** (Config Generation):
- `alacritty`: Terminal emulator
- `kitty`: Terminal emulator
- `rofi`: Application launcher
- `i3`: Window manager
- `sway`: Wayland compositor
- `dunst`: Notification daemon
- `polybar`: Status bar
- `zathura`: PDF viewer

**External Plugin Support**:
- **Language-agnostic**: Write plugins in Go, Python, Shell, Rust, or any language
- **Process isolation**: Plugins run in separate processes for stability
- **Auto-discovery**: Plugins discovered from `~/.config/tinct/plugins/` and system directories
- **HashiCorp go-plugin**: Battle-tested plugin protocol (used by Terraform, Vault)
- **Simple interfaces**: Easy to implement for custom sources or outputs

### 3. Color Extraction Algorithms
- **K-means**: Fast, vibrant palettes
- **Median Cut**: Perceptually balanced
- **Dominant Color**: Most frequent colors

### 4. Lifecycle Hooks
- **Pre-Execution**: Readiness checks, API validation, backups
- **Post-Execution**: Service reloads, wallpaper setting, notifications
- **On Error**: Rollback, cleanup, error logging

##  Technology Stack

### Core Technologies
- **Language**: Go 1.25+
- **CLI Framework**: Cobra (commands) + Viper (config)
- **Build System**: GoReleaser
- **CI/CD**: GitHub Actions

### Go 1.25+ Features Used
- **Generics**: Type-safe plugin registry
- **Range over Functions**: Custom iterators
- **Structured Logging**: log/slog package
- **Enhanced Errors**: errors.Join() for multiple errors
- **Standard Library**: slices, maps, cmp packages
- **PGO**: Profile-Guided Optimization support

### Development Tools
- **Linting**: golangci-lint (30+ linters)
- **Testing**: Built-in Go testing with race detector
- **Coverage**: Codecov integration
- **Security**: Gosec, Trivy vulnerability scanning
- **Signing**: Cosign for artifact signing
- **SBOMs**: Software Bill of Materials generation

##  Distribution

### Supported Platforms
- **Operating Systems**: Linux, macOS, Windows, FreeBSD
- **Architectures**: amd64, arm64, arm (ARMv7)

### Package Managers
- **Homebrew** (macOS/Linux): `brew install jmylchreest/tap/tinct`
- **AUR** (Arch Linux): `yay -S tinct-bin`
- **APT** (Debian/Ubuntu): `.deb` packages
- **RPM** (Fedora/RHEL): `.rpm` packages
- **APK** (Alpine): `.apk` packages
- **Nix**: Nix packages
- **Go Install**: `go install github.com/jmylchreest/tinct/cmd/tinct@latest`

##  Development Workflow

### Quick Commands
```bash
go build -o tinct ./cmd/tinct               # Build for current platform
go test ./...                                # Run tests
go test -race -cover ./...                   # Tests with race detector and coverage
golangci-lint run                            # Run linters
goreleaser release --snapshot --clean        # Create snapshot release (no tag)
goreleaser release --clean                   # Create full release (requires tag)
```

### CI/CD Pipeline
GitHub Actions workflow includes:
1. **Test Job**: Unit tests, race detection, coverage (Codecov)
2. **Lint Job**: golangci-lint, go vet, formatting checks
3. **Security Job**: Gosec, Trivy vulnerability scanning
4. **Build Job**: Snapshot builds for all platforms
5. **Release Job**: Full release with GoReleaser (tag-triggered)

### Release Process
```bash
# 1. Ensure working directory is clean
git status

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. GitHub Actions automatically:
#    - Runs all tests and security scans
#    - Builds for all platforms
#    - Creates GitHub release with changelog
#    - Updates package managers (if configured)
#    - Signs artifacts with Cosign
#    - Generates SBOMs
```

##  Documentation

### For Users
- **[README.md](README.md)**: Main user documentation
  - Installation instructions
  - Quick start guide
  - Usage examples
  - Configuration guide
  - Plugin list

- **[docs/ai-wallpaper.md](docs/ai-wallpaper.md)**: AI wallpaper generation
  - Supported AI providers
  - Setup instructions
  - Prompt engineering tips
  - Cost optimization
  - Troubleshooting

### For Developers
- **[DEVELOPMENT.md](DEVELOPMENT.md)**: Comprehensive development guidelines
  - Go 1.25+ features and best practices
  - SOLID principles in Go context
  - Idiomatic Go patterns
  - Built-in plugin development (Source & Output)
  - External plugin architecture
  - Hook system implementation
  - Testing strategies
  - Error handling patterns

- **[docs/external-plugins.md](docs/external-plugins.md)**: External plugin development
  - HashiCorp go-plugin architecture
  - Go, Python, and Shell plugin examples
  - Plugin discovery and configuration
  - Security and sandboxing
  - Publishing and distribution

- **[SETUP.md](SETUP.md)**: Project setup and configuration
  - Prerequisites
  - Initial setup steps
  - Development workflow
  - CI/CD configuration
  - Environment variables
  - Docker usage

##  Usage Examples

### Basic Usage
```bash
# Generate from existing image
tinct generate ~/wallpapers/mountain.jpg

# Generate with AI (defaults to 4K)
tinct generate --ai "cyberpunk city at night"

# Generate AI wallpaper with custom size
tinct generate --ai "mountain vista" --size ultrawide
tinct generate --ai "panorama" --size triple

# Fetch random wallpaper
tinct generate --source random-fetch --size 4k

# From color codes
tinct generate --colors "#1e1e2e,#313244,#45475a"
```

### Advanced Usage
```bash
# AI with specific provider, size, and settings
tinct generate --ai "serene mountain landscape" \
               --size 4k \
               --ai-provider stablediffusion \
               --ai-quality high \
               --ai-steps 50

# Multi-monitor setup (triple 1920x1080)
tinct generate --ai "panoramic vista" \
               --size triple \
               --outputs alacritty,kitty,rofi

# Ultrawide monitor (21:9)
tinct generate --ai "cosmic nebula" \
               --size ultrawide \
               --outputs all

# Custom dimensions
tinct generate --ai "desert sunset" \
               --size 5120x2880 \
               --preview

# Save wallpaper with specific size
tinct generate --ai "forest scene" \
               --size dual-4k \
               --save-wallpaper ~/Pictures/wallpaper.png
```

##  Security Features

- **Static Analysis**: Gosec security scanning
- **Dependency Scanning**: Trivy vulnerability detection
- **Artifact Signing**: Cosign keyless signing
- **SBOM Generation**: Complete dependency tracking
- **Secure Defaults**: No hardcoded secrets, API keys via environment

##  Quality Metrics

### Testing
- **Coverage Target**: >80% for critical paths, 100% for public APIs
- **Test Types**: Unit, integration, table-driven, benchmarks
- **Race Detection**: All tests run with `-race` flag
- **Continuous Testing**: GitHub Actions on every push/PR

### Code Quality
- **Linting**: 30+ linters via golangci-lint
- **Formatting**: gofmt, goimports
- **Static Analysis**: go vet, staticcheck
- **Complexity**: gocyclo, gocognit monitoring

##  Design Principles

### SOLID Principles
- **Single Responsibility**: Each type/package has one clear purpose
- **Open/Closed**: Extensible via interfaces, no modification needed
- **Liskov Substitution**: Interface implementations honor contracts
- **Interface Segregation**: Small, focused interfaces
- **Dependency Inversion**: Depend on abstractions, not concretions

### Go Best Practices
- Accept interfaces, return structs
- Use context for cancellation
- Wrap errors with context
- Defer for cleanup
- Table-driven tests
- Functional options pattern

##  Resources

### Go Language
- [Go 1.25 Language Specification](https://go.dev/ref/spec)
- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Effective Go](https://go.dev/doc/effective_go)

### Project Links
- **Repository**: https://github.com/jmylchreest/tinct
- **Issues**: https://github.com/jmylchreest/tinct/issues
- **Discussions**: https://github.com/jmylchreest/tinct/discussions

### Inspiration
- [pywal](https://github.com/dylanaraps/pywal): System theming from images
- [mutagen](https://mutagen.io/): Plugin architecture and lifecycle hooks
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin): External plugin system
- [Cobra](https://github.com/spf13/cobra): CLI framework
- [Viper](https://github.com/spf13/viper): Configuration management

##  Contributing

1. Read [DEVELOPMENT.md](DEVELOPMENT.md)
2. Fork the repository
3. Create feature branch: `git checkout -b feat/amazing-feature`
4. Follow conventional commits: `feat:`, `fix:`, `docs:`, etc.
5. Run tests and linting: `go test ./... && golangci-lint run`
6. Push and create pull request

##  License

MIT License - See [LICENSE](LICENSE) file

// Copyright (c) 2025 John Mylchreest

---

**Project Status**:  In Development

**Next Steps**:
1. Implement core colour extraction algorithms
2. Build plugin system (Source & Output)
3. Implement hook system
4. Create built-in plugins (Alacritty, Kitty, etc.)
5. Integrate AI providers (Stable Diffusion, DALL-E)
6. Write comprehensive tests
7. Create documentation and examples
8. First release (v1.0.0)

---

**Made with  and Go 1.25+**