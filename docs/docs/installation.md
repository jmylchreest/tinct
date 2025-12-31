---
sidebar_position: 2
---

# Installation

Tinct can be installed through several methods depending on your distribution and preferences.

## Arch Linux (AUR)

The recommended method for Arch Linux users:

```bash
# Using yay
yay -S tinct-bin

# Using paru
paru -S tinct-bin
```

## From Go

If you have Go 1.25+ installed:

```bash
go install github.com/jmylchreest/tinct/cmd/tinct@latest
```

Ensure `$GOPATH/bin` (typically `~/go/bin`) is in your `PATH`.

## From source

Clone and build manually:

```bash
git clone https://github.com/jmylchreest/tinct.git
cd tinct
go build -o tinct ./cmd/tinct

# Optionally install to your PATH
sudo mv tinct /usr/local/bin/
```

## Verify installation

After installation, verify tinct is working:

```bash
tinct --version
```

You should see version information displayed.

## Shell completions

Tinct supports shell completions for Bash, Zsh, Fish, and PowerShell:

```bash
# Bash (add to ~/.bashrc)
source <(tinct completion bash)

# Zsh (add to ~/.zshrc)
source <(tinct completion zsh)

# Fish
tinct completion fish | source

# PowerShell
tinct completion powershell | Out-String | Invoke-Expression
```

## Optional dependencies

Some features require additional tools:

| Feature | Dependency | Purpose |
|---------|------------|---------|
| AI image generation | `GOOGLE_API_KEY` | Google Gemini image generation |
| OpenRouter integration | `OPENROUTER_API_KEY` | Multi-model AI generation |
| GNOME Shell theming | User Themes extension | Automatic theme application |
| GTK3 theming | adw-gtk3-theme | GTK3 theme support |
| Qt5/Qt6 theming | qt5ct/qt6ct | Qt application theming |
| KDE Plasma | plasma-apply-colorscheme | Color scheme application |

## Configuration directory

Tinct creates configuration files in `~/.config/tinct/`:

- `.tinct-manifest.json` - Tracks generated files
- `.tinct-plugins.json` - Plugin configuration
- `themes/` - Saved theme files (when using markdown output)

## Next steps

- [Quick start](/docs/quickstart) - Generate your first theme
- [Plugin management](/docs/commands/plugins) - Configure and install plugins
