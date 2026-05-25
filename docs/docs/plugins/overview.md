---
sidebar_position: 1
---

# Plugin architecture

Tinct uses a plugin-based architecture for both input (colour sources) and output (application configurations).

## Plugin types

### Input plugins

Extract or generate colour palettes from various sources:

| Plugin | Description |
|--------|-------------|
| [image](./input/image.md) | Extract from JPEG, PNG, GIF, WebP images |
| [google-genai](./input/google-genai) | Generate with Google Gemini AI |
| [openrouter](./input/openrouter) | Generate via OpenRouter API |
| [remote-json](./input/remote-json.md) | Fetch from JSON URLs |
| [remote-css](./input/remote-css.md) | Parse CSS variables |
| [file](./input/file.md) | Manual colour specification |
| [markdown](./input/markdown.md) | Load saved theme files |

### Output plugins

Generate configuration files for applications:

- [Terminals](./output/index.md#terminals) - Alacritty, Ghostty, Kitty, Konsole, Ptyxis
- [Desktop environments](./output/index.md#desktop-environments) - GNOME, KDE, GTK, Qt
- [Window managers](./output/index.md#hyprland-ecosystem) - Hyprland, Hyprlock, Hyprpaper
- [Bars and launchers](./output/index.md#bars-and-launchers) - Waybar, Dunst, Fuzzel, etc.
- [Editors](./output/index.md#editors-and-multiplexers) - Neovim, Zellij
- [Special](./output/index.md#special-purpose) - Markdown export, custom templates

## Built-in vs external

### Built-in plugins

- Compiled into the tinct binary
- Written in Go
- Maximum performance
- No external dependencies

### External plugins

- Separate executables
- Written in any language
- Two protocols: JSON-stdio and go-plugin
- Loaded on demand

See [Plugin development](../plugin-development/index.md) for details.

## Plugin lifecycle

1. **Registration**: Plugins register with the manager on startup
2. **Flag registration**: Plugins add their CLI flags
3. **Validation**: Pre-execution checks (dependencies, paths)
4. **Execution**: Generate method is called with theme data
5. **Post-hooks**: Optional reload/apply actions

## Using plugins

### Specify input plugin

```bash
tinct generate -i <plugin-name>
```

### Specify output plugins

```bash
# Single plugin
tinct generate -i image -p wallpaper.jpg -o kitty

# Multiple plugins
tinct generate -i image -p wallpaper.jpg -o kitty,hyprland,waybar

# All plugins
tinct generate -i image -p wallpaper.jpg -o all
```

### Plugin-specific flags

Plugins can define additional flags:

```bash
tinct generate -i image -p wallpaper.jpg \
  --image.colors 16 \
  --image.extractAmbience \
  --kitty.reload
```

## Listing plugins

See all available plugins:

```bash
tinct plugins list
```

## Next steps

- [Input plugins](./input/index.md) - Colour source plugins
- [Output plugins](./output/index.md) - Application plugins
- [Plugin development](../plugin-development/index.md) - Creating custom plugins
