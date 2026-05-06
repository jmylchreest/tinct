---
sidebar_position: 1
---

# Commands

Tinct provides several commands for generating themes, extracting palettes, and managing plugins.

## Available commands

| Command | Description |
|---------|-------------|
| [`generate`](./generate.md) | Generate theme configurations from a colour source |
| [`extract`](./extract.md) | Extract colours without generating configurations |
| [`plugins`](./plugins.md) | Manage input and output plugins |
| [`files`](./files.md) | Track and manage generated files |

## Global flags

These flags are available on all commands:

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Enable verbose output |
| `-q, --quiet` | Suppress non-error output |
| `-h, --help` | Show help for the command |
| `--version` | Show version information |

## Command structure

```bash
tinct <command> [subcommand] [flags]
```

### Examples

```bash
# Generate theme from image
tinct generate -i image -p ~/wallpaper.jpg -o all

# Extract palette only
tinct extract -i image -p ~/wallpaper.jpg

# List plugins
tinct plugins list

# Show tracked files
tinct files list
```

## Getting help

Get help for any command:

```bash
# General help
tinct --help

# Command-specific help
tinct generate --help
tinct plugins --help
```

## Shell completion

Enable tab completion for your shell:

```bash
# Bash
source <(tinct completion bash)

# Zsh
source <(tinct completion zsh)

# Fish
tinct completion fish | source
```
