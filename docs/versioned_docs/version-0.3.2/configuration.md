---
sidebar_position: 4
---

# Configuration

Tinct reads its persistent settings from `~/.config/tinct/tinct.toml` (or
`$XDG_CONFIG_HOME/tinct/tinct.toml` if set). The file is created on first
run with sensible defaults; you can edit it freely.

## Per-plugin output directories

Each output plugin writes generated files to a sensible default location
(documented in each plugin's reference page). If you'd rather files land
somewhere else — say, a synced dotfiles directory — you have three ways
to override:

1. **CLI flag** for one-off use: `--<plugin>.output-dir=/path`
2. **Environment variable** for shells/scripts: `TINCT_PLUGIN_<NAME>_OUTPUT_DIR=/path`
3. **`tinct.toml`** for persistent preferences: `[plugin.<name>]` table

### Example

```toml
# ~/.config/tinct/tinct.toml

[plugin.kitty]
output_dir = "/home/me/dotfiles/kitty/themes"

[plugin.neovim]
output_dir = "/home/me/.config/nvim/lua/themes"

[plugin.alacritty]
output_dir = "/home/me/dotfiles/alacritty"
```

The plugin name is the same as in `tinct generate -o <name>` — for
example `[plugin.kde-plasma]`, `[plugin.gnome-shell]`, `[plugin.gtk3]`.
Hyphens in plugin names are kept; the env var equivalent uppercases and
substitutes them with underscores:

| Plugin | TOML | Env var |
|---|---|---|
| `kitty` | `[plugin.kitty] output_dir` | `TINCT_PLUGIN_KITTY_OUTPUT_DIR` |
| `kde-plasma` | `[plugin.kde-plasma] output_dir` | `TINCT_PLUGIN_KDE_PLASMA_OUTPUT_DIR` |
| `google-genai` | `[plugin.google-genai] output_dir` | `TINCT_PLUGIN_GOOGLE_GENAI_OUTPUT_DIR` |

### Precedence

Highest priority wins. From the user's perspective:

1. **CLI flag** (`--kitty.output-dir=/foo`) — always overrides everything.
2. **Environment variable** (`TINCT_PLUGIN_KITTY_OUTPUT_DIR=/foo`) — useful in shell scripts and CI.
3. **`tinct.toml`** (`[plugin.kitty] output_dir = "/foo"`) — your saved default.
4. **Plugin's built-in default** (e.g., `~/.config/kitty/themes`) — what you get out of the box.

If you set the env var but pass `--kitty.output-dir=/somewhere/else`,
the flag wins. If neither flag nor env is set, tinct.toml wins. If none
of those are set, the plugin's platform-resolved default is used (XDG
config dir on Linux/macOS, `%APPDATA%` on Windows).

## Other settings

Telemetry and image cache settings live under their own tables:

```toml
[telemetry]
# Set to false to disable anonymous telemetry. Default: true.
enabled = true

[cache]
# Cache remote images locally. Default: true.
images = true

# Override the cache directory. Default: $XDG_CACHE_HOME/tinct.
dir = ""

# Re-download cached images on every run. Default: false.
overwrite = false
```

These also support env-var overrides (`TINCT_TELEMETRY=off`,
`TINCT_IMAGE_CACHE=false`, `TINCT_IMAGE_CACHE_DIR=/path`,
`TINCT_IMAGE_CACHE_OVERWRITE=true`).

## File location

| Platform | Path |
|---|---|
| Linux | `$XDG_CONFIG_HOME/tinct/tinct.toml` (default `~/.config/tinct/tinct.toml`) |
| macOS | `~/.config/tinct/tinct.toml` |
| Windows | `%APPDATA%/tinct/tinct.toml` |

The file is plain TOML; any text editor works. Tinct re-reads the file on
each run.
