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

```bash
# Extract and preview colours from an image
tinct extract -i image -p ~/Pictures/wallpaper.jpg

# Apply theme to all configured applications
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all

# Use a popular theme (Catppuccin Mocha)
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o all
```

## Features

- **Multiple input sources** - Images, AI generation, remote themes (JSON/CSS), manual specification
- **25+ applications** - Terminals, desktops, window managers, bars, editors
- **Smart categorisation** - Auto-assigns semantic colours with WCAG contrast checking
- **Plugin architecture** - Extend with custom inputs/outputs in any language
- **Theme portability** - Save and share complete themes as markdown files

## Supported applications

| Category | Applications |
|----------|-------------|
| Terminals | Alacritty, Ghostty, Kitty, Konsole, Ptyxis |
| Desktop | GNOME Shell, KDE Plasma, GTK3/4, Libadwaita, Qt5/6 |
| Hyprland | Hyprland, Hyprpaper, Hyprlock |
| Bars & Launchers | Waybar, Dunst, SwayOSD, Fuzzel, Walker, Wofi |
| Editors | Neovim, Zellij |

## Documentation

Full documentation is available at **[jmylchreest.github.io/tinct](https://jmylchreest.github.io/tinct/)**

- [Getting started](https://jmylchreest.github.io/tinct/docs/quickstart)
- [Plugin reference](https://jmylchreest.github.io/tinct/docs/plugins/overview)
- [Templating guide](https://jmylchreest.github.io/tinct/docs/templating)
- [Plugin development](https://jmylchreest.github.io/tinct/docs/plugin-development)

## Contributing

Contributions welcome! See the [documentation](https://jmylchreest.github.io/tinct/) for development guides.

## License

MIT License - see [LICENSE](LICENSE)
