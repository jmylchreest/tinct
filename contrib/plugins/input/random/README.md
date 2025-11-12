# Random Colour Palette Generator

A Tinct input plugin that generates random colour palettes with configurable seed and colour count.

## Features

- Generate random colour palettes with configurable colour counts
- Deterministic generation with seed support for reproducibility
- Uses go-plugin RPC protocol for better performance and process isolation
- Process reuse across multiple invocations
- Dry-run mode support
- Verbose output option

## Installation

### From Release

Download the latest release for your platform:

```bash
# Linux x86_64
tinct plugins add https://github.com/jmylchreest/tinct/releases/latest/download/tinct-plugin-random_VERSION_Linux_x86_64.tar.gz:random

# macOS (Darwin) x86_64
tinct plugins add https://github.com/jmylchreest/tinct/releases/latest/download/tinct-plugin-random_VERSION_Darwin_x86_64.tar.gz:random

# macOS (Darwin) ARM64
tinct plugins add https://github.com/jmylchreest/tinct/releases/latest/download/tinct-plugin-random_VERSION_Darwin_arm64.tar.gz:random
```

Replace `VERSION` with the actual version number.

### From Source

```bash
cd contrib/plugins/input/random
go build -o tinct-plugin-random
tinct plugins add ./random --type input
```

## Usage

### Basic Usage

Generate a palette with 32 random colours (default):

```bash
tinct generate -i random -o tailwind
```

### Custom Colour Count

Generate a palette with a specific number of colours:

```bash
tinct generate -i random -o tailwind --plugin-arg count=16
```

### Reproducible Generation

Use a seed for deterministic colour generation:

```bash
# Generate with specific seed
tinct generate -i random -o tailwind --plugin-arg seed=12345

# Running again with the same seed produces identical colours
tinct generate -i random -o tailwind --plugin-arg seed=12345
```

### Combined Options

```bash
tinct generate -i random -o tailwind \
  --plugin-arg count=24 \
  --plugin-arg seed=42
```

## Plugin Arguments

| Argument | Type | Default | Description |
|----------|------|---------|-------------|
| `count` | integer | 32 | Number of colours to generate |
| `seed` | integer | random | Random seed for reproducible generation. If not provided, a cryptographically random seed is used |

## Configuration Example

Add to your `tinct.yaml`:

```yaml
inputs:
  - name: random
    type: external
    path: /path/to/random
    args:
      count: 32
      seed: 12345  # Optional: remove for random generation
```

## Dry Run

Test the plugin without generating actual output:

```bash
tinct generate -i random -o tailwind --dry-run --verbose
```

Output:
```
DRY-RUN MODE: Would generate 32 random colours
Random seed: 1234567890
```

## Technical Details

- **Protocol**: go-plugin RPC (HashiCorp's plugin system)
- **Plugin Type**: Input
- **Protocol Version**: 1
- **Language**: Go 1.21+

## Building

Requirements:
- Go 1.21 or higher
- Access to Tinct's internal packages

Build command:
```bash
go build -o tinct-plugin-random
```

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please submit issues and pull requests to the main Tinct repository.
