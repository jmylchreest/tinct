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
| [image](/docs/plugins/input/image) | Extract from JPEG, PNG, GIF, WebP images |
| [google-genai](/docs/plugins/input/google-genai) | Generate with Google Gemini AI |
| [openrouter](/docs/plugins/input/openrouter) | Generate via OpenRouter API |
| [remote-json](/docs/plugins/input/remote-json) | Fetch from JSON URLs |
| [remote-css](/docs/plugins/input/remote-css) | Parse CSS variables |
| [file](/docs/plugins/input/file) | Manual colour specification |
| [markdown](/docs/plugins/input/markdown) | Load saved theme files |

### Output plugins

Generate configuration files for applications:

- [Terminals](/docs/plugins/output#terminals) - Alacritty, Ghostty, Kitty, Konsole, Ptyxis
- [Desktop environments](/docs/plugins/output#desktop-environments) - GNOME, KDE, GTK, Qt
- [Window managers](/docs/plugins/output#hyprland-ecosystem) - Hyprland, Hyprlock, Hyprpaper
- [Bars and launchers](/docs/plugins/output#bars-and-launchers) - Waybar, Dunst, Fuzzel, etc.
- [Editors](/docs/plugins/output#editors-and-multiplexers) - Neovim, Zellij
- [Special](/docs/plugins/output#special-purpose) - Markdown export, custom templates

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

See [Plugin development](/docs/plugin-development) for details.

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

- [Input plugins](/docs/plugins/input) - Colour source plugins
- [Output plugins](/docs/plugins/output) - Application plugins
- [Plugin development](/docs/plugin-development) - Creating custom plugins
