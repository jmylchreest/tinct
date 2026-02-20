---
sidebar_position: 1
---

# Plugin development

Create custom plugins to extend tinct's functionality.

## Overview

Tinct's plugin architecture lets you:

- **Input plugins**: Add new colour sources (APIs, file formats, devices)
- **Output plugins**: Generate configs for any application or device

Plugins can be written in any language using the JSON-stdio protocol, or in Go using the high-performance go-plugin RPC protocol.

## Getting started

1. **Choose a protocol**
   - [JSON-stdio](/docs/plugin-development/protocols#json-stdio-protocol) - Any language, simple to implement
   - [Go-plugin RPC](/docs/plugin-development/protocols#go-plugin-protocol) - Go only, better performance

2. **Create your plugin**
   - [Shell script example](/docs/plugin-development/creating#shell-script-plugin)
   - [Python example](/docs/plugin-development/creating#python-plugin)
   - [Go example (JSON-stdio)](/docs/plugin-development/creating#go-plugin-json-stdio)
   - [Go example (RPC)](/docs/plugin-development/creating#go-plugin-go-plugin-rpc)

3. **Add lifecycle hooks** (optional)
   - [PreExecute](/docs/plugin-development/hooks#preexecutehook-interface) - Validate environment
   - [PostExecute](/docs/plugin-development/hooks#postexecutehook-interface) - Reload configs

4. **Publish your plugin**
   - [Repository format](/docs/plugin-development/publishing#repository-structure)
   - [Release workflow](/docs/plugin-development/publishing#creating-releases)

## Plugin types

### Input plugins

Input plugins provide colour palettes to tinct. They receive configuration and return a palette:

```
┌─────────────┐    config     ┌──────────────┐    palette    ┌─────────┐
│    tinct    │ ───────────▶  │ input plugin │ ───────────▶  │  tinct  │
└─────────────┘               └──────────────┘               └─────────┘
```

Examples: image extraction, AI generation, remote API fetching

### Output plugins

Output plugins receive a palette and generate configuration files:

```
┌─────────────┐    palette    ┌───────────────┐    files     ┌─────────┐
│    tinct    │ ───────────▶  │ output plugin │ ───────────▶  │  disk   │
└─────────────┘               └───────────────┘               └─────────┘
```

Examples: terminal configs, desktop themes, LED controllers

## Plugin naming

Follow this naming convention:

```
tinct-{type}-{name}
```

Examples:
- `tinct-input-spotify` - Input plugin for Spotify album art
- `tinct-output-wled` - Output plugin for WLED LED strips
- `tinct-output-openrgb` - Output plugin for OpenRGB

## Installation paths

Plugins are discovered from:

1. `~/.config/tinct/plugins/` - User plugins (highest priority)
2. `~/.local/share/tinct/plugins/` - Installed from repositories
3. System paths defined by `$TINCT_PLUGIN_PATH`

## Quick example

A minimal shell script output plugin:

```bash
#!/bin/bash
# tinct-output-example

if [[ "$1" == "--plugin-info" ]]; then
  echo '{"name":"example","type":"output","version":"1.0.0","protocol_version":"0.0.1","description":"Example plugin","plugin_protocol":"json-stdio"}'
  exit 0
fi

# Read palette, extract background colour, write config
jq -r '"background=\(.colours.background.hex)"' > ~/.config/myapp/theme.conf
```

## Next steps

- [Protocol details](/docs/plugin-development/protocols) - Understand communication
- [Creating plugins](/docs/plugin-development/creating) - Full examples
- [Lifecycle hooks](/docs/plugin-development/hooks) - Advanced control
