# Tinct

> An extensible colour palette generator and theme manager for unified theming across your entire environment

Generate colour palettes from images, AI, remote themes, or manual specifications. Apply them consistently to terminals, desktop environments, window managers, and external devices.

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-jmylchreest.github.io%2Ftinct-blue)](https://jmylchreest.github.io/tinct/)

## Installation

**Arch Linux (AUR):**
```bash
yay -S tinct-bin
```

**From Go:**
```bash
go install github.com/jmylchreest/tinct/cmd/tinct@latest
```

**From source:**
```bash
git clone https://github.com/jmylchreest/tinct.git
cd tinct && go build -o tinct ./cmd/tinct
```

## Quick start

### From an image

```bash
# Apply theme to all configured applications
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all
```

### From AI (Google Gemini)

Generate a unique wallpaper and theme from a text prompt:

```bash
# Set your API key (get one at https://aistudio.google.com/apikey)
export GOOGLE_API_KEY="your-api-key"

# Or retrieve it from a secret manager
export GOOGLE_API_KEY=$(secret-tool lookup service google-genai)  # GNOME Keyring
export GOOGLE_API_KEY=$(bw get password google-genai)             # Bitwarden CLI

# Generate a wallpaper and apply the extracted palette
tinct generate -i google-genai --ai.prompt "sunset over rolling hills of tuscany" -o all
```

OpenRouter is also supported as an alternative (`-i openrouter`), giving access to multiple AI models. See the [AI generation docs](https://jmylchreest.github.io/tinct/docs/plugins/input/google-genai) for details.

### From a popular theme

```bash
# Catppuccin Mocha
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o all
```

## Features

- **Multiple input sources** - Images, AI generation (Google Gemini, OpenRouter), remote themes (JSON/CSS), manual specification
- **25+ applications** - Terminals, desktops, window managers, bars, editors
- **Smart categorisation** - Auto-assigns semantic colours with WCAG contrast checking
- **Plugin ecosystem** - Extend with external plugins from repositories or build your own in any language
- **Theme portability** - Save and share complete themes as markdown files

## Supported applications

| Category | Applications |
|----------|-------------|
| Terminals | Alacritty, Ghostty, Kitty, Konsole, Ptyxis |
| Desktop | GNOME Shell, KDE Plasma, GTK3/4, Libadwaita, Qt5/6 |
| Hyprland | Hyprland, Hyprpaper, Hyprlock |
| Bars & Launchers | Waybar, Dunst, SwayOSD, Fuzzel, Walker, Wofi |
| Editors | Neovim, Zellij |

Need more? Install external plugins from the [plugin repository](https://github.com/jmylchreest/tinct-plugins):

```bash
tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/refs/heads/main/repository/repository.json
tinct plugins search
tinct plugins install <plugin-name>
```

## Documentation

Full documentation is available at **[jmylchreest.github.io/tinct](https://jmylchreest.github.io/tinct/)**

- [Getting started](https://jmylchreest.github.io/tinct/docs/quickstart)
- [Plugin reference](https://jmylchreest.github.io/tinct/docs/plugins/overview)
- [Templating guide](https://jmylchreest.github.io/tinct/docs/templating)
- [Plugin development](https://jmylchreest.github.io/tinct/docs/plugin-development)

## Telemetry

Since v0.1.25, tinct collects anonymous usage telemetry to help prioritise plugin development and maintenance. Only basic event data is sent: OS name, app version, plugin names/versions used, and feature flags (theme type, seed mode, backend, dry-run, etc.). No personal information, file paths, prompts, API keys, or colour data is ever collected. Telemetry can be disabled by setting `TINCT_TELEMETRY=off` or `{"enabled": false}` in `~/.local/share/tinct/telemetry.json`. See the [telemetry docs](https://jmylchreest.github.io/tinct/docs/telemetry) for full details.

## Contributing

Contributions welcome! See the [documentation](https://jmylchreest.github.io/tinct/) for development guides.

## License

MIT License - see [LICENSE](LICENSE)
