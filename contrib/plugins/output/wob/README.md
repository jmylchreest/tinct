# tinct-plugin-wob - Wob Theme Generator and Wrapper

A dual-mode tool that integrates [wob](https://github.com/francma/wob) (Wayland Overlay Bar) with Tinct, providing both automatic theme generation and a convenient wrapper for managing wob instances.

## Features

- **Tinct Plugin Mode**: Automatically generates wob themes from Tinct colour palettes
- **Wrapper Mode**: Manages wob instances (start, stop, send values)
- **Self-Installing**: Plugin mode installs the wrapper script automatically
- **Single Binary**: No dependencies, pure Go implementation
- **Runtime Safety**: Secure FIFO handling with proper permissions
- **Hyprland Integration**: Ready-to-use examples for volume/brightness controls

## Installation

### From Source

```bash
cd contrib/plugins/output/wob
go build -o tinct-plugin-wob
```

### As Tinct Plugin

1. Copy the binary to your Tinct plugins directory:

```bash
mkdir -p ~/.config/tinct/plugins
cp tinct-plugin-wob ~/.config/tinct/plugins/
```

2. The plugin will be automatically discovered when you use:

```bash
tinct generate -i image -p wallpaper.jpg -o wob
```

## Usage

### Plugin Mode (Used by Tinct)

When Tinct generates themes with `-o wob`, the plugin:

1. **Generates theme file**: `~/.config/wob/themes/tinct.ini`
2. **Installs wrapper**: `~/.config/wob/scripts/tinct-plugin-wob` (copy of itself)
3. **Provides instructions**: Shows Hyprland integration examples

**Example:**

```bash
# Generate wob theme from your wallpaper
tinct generate -i image -p ~/wallpaper.jpg -o wob

# Output:
# Generated wob theme: ~/.config/wob/themes/tinct.ini
# Installed wrapper: ~/.config/wob/scripts/tinct-plugin-wob
#
# To use with Hyprland, add to your hyprland.conf:
#   exec-once = ~/.config/wob/scripts/tinct-plugin-wob start ...
```

### Wrapper Mode (Direct Usage)

#### Start wob with theme

```bash
# Start with base config only
tinct-plugin-wob start --base-config ~/.config/wob/base.ini

# Start with base + theme overlay
tinct-plugin-wob start --base-config ~/.config/wob/base.ini \
                --append-config ~/.config/wob/themes/tinct.ini

# Start with multiple theme layers
tinct-plugin-wob start --base-config ~/.config/wob/base.ini \
                --append-config ~/.config/wob/themes/tinct.ini \
                --append-config ~/.config/wob/custom-overrides.ini
```

#### Send values

```bash
# Send a percentage value (0-100)
tinct-plugin-wob send 45

# Send current/max (automatically calculates percentage)
tinct-plugin-wob send 4234 9600  # Sends: 44%

# Send with style override
tinct-plugin-wob send 95 --style critical
tinct-plugin-wob send 50 --style warning
tinct-plugin-wob send 30 --style normal
```

#### Manage wob

```bash
# Check if wob is running
tinct-plugin-wob status

# Stop wob
tinct-plugin-wob stop

# Show version
tinct-plugin-wob version
```

## Hyprland Integration

### Basic Setup

Add to your `~/.config/hypr/hyprland.conf`:

```conf
# Start wob with Tinct theme on login
exec-once = ~/.config/wob/scripts/tinct-plugin-wob start \
    --base-config ~/.config/wob/base.ini \
    --append-config ~/.config/wob/themes/tinct.ini
```

### Volume Controls

```conf
# Volume up
bind = , XF86AudioRaiseVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%+ && \
    ~/.config/wob/scripts/tinct-plugin-wob send $(wpctl get-volume @DEFAULT_SINK@ | awk '{print int($2 * 100)}')

# Volume down
bind = , XF86AudioLowerVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%- && \
    ~/.config/wob/scripts/tinct-plugin-wob send $(wpctl get-volume @DEFAULT_SINK@ | awk '{print int($2 * 100)}')

# Mute toggle (sends 0)
bind = , XF86AudioMute, exec, wpctl set-mute @DEFAULT_SINK@ toggle && \
    ~/.config/wob/scripts/tinct-plugin-wob send 0
```

### Brightness Controls

```conf
# Brightness up
bind = , XF86MonBrightnessUp, exec, brightnessctl set 5%+ && \
    ~/.config/wob/scripts/tinct-plugin-wob send $(brightnessctl get) $(brightnessctl max)

# Brightness down
bind = , XF86MonBrightnessDown, exec, brightnessctl set 5%- && \
    ~/.config/wob/scripts/tinct-plugin-wob send $(brightnessctl get) $(brightnessctl max)
```

### With Style-Based Thresholds

```conf
# Volume with automatic critical style at >80%
bind = , XF86AudioRaiseVolume, exec, \
    wpctl set-volume @DEFAULT_SINK@ 5%+ && \
    vol=$(wpctl get-volume @DEFAULT_SINK@ | awk '{print int($2 * 100)}') && \
    [[ $vol -gt 80 ]] && style="critical" || style="normal" && \
    ~/.config/wob/scripts/tinct-plugin-wob send $vol --style $style
```

## Configuration

### Base Configuration

Create `~/.config/wob/base.ini`:

```ini
[default]
# Bar dimensions
height = 50
width = 400
border_offset = 4
border_size = 2
bar_padding = 3
anchor = top
margin = 20

# Timeout (milliseconds)
timeout = 2000

# Colours will be overridden by theme
background_color = FF000000
border_color = FFFFFFFF
bar_color = FFFFFFFF
```

### Generated Theme

Tinct generates `~/.config/wob/themes/tinct.ini` with:

- **default** section: Base colours from your palette
- **normal** section: Success/positive colour (green)
- **critical** section: Danger/error colour (red)
- **warning** section: Warning/caution colour (yellow/orange)

Example generated theme:

```ini
# Wob theme generated by Tinct

[default]
height = 50
border_offset = 4
border_size = 2
bar_padding = 3

background_color = FF1e1e2e
border_color = FFcdd6f4
bar_color = FF89b4fa

[normal]
bar_color = FFa6e3a1

[critical]
bar_color = FFf38ba8

[warning]
bar_color = FFf9e2af
```

### Custom Overrides

Create additional config files to override specific settings:

`~/.config/wob/custom.ini`:

```ini
[default]
# Make the bar taller
height = 60

# Position at bottom instead
anchor = bottom
margin = 30
```

Then load with multiple `--append-config` flags (last config wins):

```bash
tinct-plugin-wob start --base-config base.ini \
                --append-config themes/tinct.ini \
                --append-config custom.ini
```

## Runtime Behavior

### Directory Structure

```
$XDG_RUNTIME_DIR/wob/  (or /tmp/$USER-wob-runtime/)
├── wob.fifo           # Named pipe for sending values
├── wob-merged.ini     # Merged configuration (if using multiple configs)
└── wob.pid            # PID file for process management
```

### Security

- Runtime directory: `0700` permissions (user-only access)
- FIFO: `0600` permissions (user-only read/write)
- Ownership verification on startup
- Symlink attack protection

### Auto-Start Behavior

When you send a value but wob isn't running:

```bash
tinct-plugin-wob send 50  # Automatically starts wob with default config
```

To start with your theme automatically, ensure wob is started via `exec-once` in your Hyprland config.

## Environment Variables

- `WOB_BIN`: Override wob binary (default: `wob`)
- `WOB_RUNTIME_DIR`: Override runtime directory
- `WOB_PIPE`: Override FIFO path
- `WOB_MERGED_CONFIG`: Override merged config path

Example:

```bash
export WOB_BIN=/usr/local/bin/wob
export WOB_RUNTIME_DIR=/run/user/$UID/my-wob
tinct-plugin-wob start --base-config base.ini
```

## Troubleshooting

### "wob is already running"

If you see this message but wob isn't actually running, remove the stale PID file:

```bash
rm -f $XDG_RUNTIME_DIR/wob/wob.pid
```

### "failed to open FIFO"

Ensure the FIFO exists and has correct permissions:

```bash
ls -la $XDG_RUNTIME_DIR/wob/
# Should show: prw------- wob.fifo
```

If missing or incorrect, stop and restart wob:

```bash
tinct-plugin-wob stop
tinct-plugin-wob start --base-config ~/.config/wob/base.ini
```

### Values not displaying

1. Check if wob is running:

```bash
tinct-plugin-wob status
```

2. Test with a direct value:

```bash
tinct-plugin-wob send 50
```

3. Check wob process:

```bash
ps aux | grep wob
cat $XDG_RUNTIME_DIR/wob/wob.pid
```

### Config not applying

Ensure config merge order is correct (last wins):

```bash
tinct-plugin-wob start --base-config base.ini \
                --append-config themes/tinct.ini  # This overrides base.ini
```

## Examples

### Complete Hyprland Setup

`~/.config/hypr/hyprland.conf`:

```conf
# Start wob with Tinct theme
exec-once = ~/.config/wob/scripts/tinct-plugin-wob start \
    --base-config ~/.config/wob/base.ini \
    --append-config ~/.config/wob/themes/tinct.ini

# Volume controls
bind = , XF86AudioRaiseVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%+ && \
    vol=$(wpctl get-volume @DEFAULT_SINK@ | awk '{print int($2 * 100)}') && \
    ~/.config/wob/scripts/tinct-plugin-wob send $vol

bind = , XF86AudioLowerVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%- && \
    vol=$(wpctl get-volume @DEFAULT_SINK@ | awk '{print int($2 * 100)}') && \
    ~/.config/wob/scripts/tinct-plugin-wob send $vol

bind = , XF86AudioMute, exec, wpctl set-mute @DEFAULT_SINK@ toggle && \
    ~/.config/wob/scripts/tinct-plugin-wob send 0

# Brightness controls
bind = , XF86MonBrightnessUp, exec, brightnessctl set 5%+ && \
    ~/.config/wob/scripts/tinct-plugin-wob send $(brightnessctl get) $(brightnessctl max)

bind = , XF86MonBrightnessDown, exec, brightnessctl set 5%- && \
    ~/.config/wob/scripts/tinct-plugin-wob send $(brightnessctl get) $(brightnessctl max)

# Reload theme after generating new palette
bind = $mainMod SHIFT, R, exec, \
    tinct generate -i image -p ~/wallpaper.jpg -o wob,hyprland && \
    ~/.config/wob/scripts/tinct-plugin-wob stop && \
    sleep 0.1 && \
    ~/.config/wob/scripts/tinct-plugin-wob start \
        --base-config ~/.config/wob/base.ini \
        --append-config ~/.config/wob/themes/tinct.ini
```

### Script-Based Integration

Create a helper script at `~/.local/bin/wob-volume`:

```bash
#!/bin/bash
# wob-volume - Send volume to wob with style based on level

vol=$(wpctl get-volume @DEFAULT_SINK@ | awk '{print int($2 * 100)}')

if [[ $vol -gt 80 ]]; then
    style="critical"
elif [[ $vol -gt 60 ]]; then
    style="warning"
else
    style="normal"
fi

~/.config/wob/scripts/tinct-plugin-wob send $vol --style $style
```

Then use in Hyprland:

```conf
bind = , XF86AudioRaiseVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%+ && wob-volume
bind = , XF86AudioLowerVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%- && wob-volume
```

## Architecture

### Dual-Mode Design

```
tinct-plugin-wob binary
├── Plugin Mode (invoked by Tinct)
│   ├── Read palette JSON from stdin
│   ├── Generate ~/.config/wob/themes/tinct.ini
│   ├── Install self to ~/.config/wob/scripts/tinct-plugin-wob
│   └── Print Hyprland integration instructions
│
└── Wrapper Mode (invoked by user/scripts)
    ├── start: Launch wob with merged config
    ├── send: Write values to FIFO
    ├── stop: Terminate wob process
    └── status: Check if running
```

### Why Single Binary?

1. **Simplicity**: One file to distribute and manage
2. **Self-contained**: Plugin installs its own wrapper
3. **Type safety**: Go instead of bash string manipulation
4. **Better error handling**: Structured errors and validation
5. **Cross-platform**: Compiles for Linux/BSD/etc.

## See Also

- [wob](https://github.com/francma/wob) - Wayland Overlay Bar
- [Tinct](https://github.com/jmylchreest/tinct) - Colour palette generator
- [Hyprland](https://hyprland.org/) - Dynamic tiling Wayland compositor

## License

See main Tinct repository for license information.
