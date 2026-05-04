---
sidebar_position: 1
---

# Input plugins

Input plugins extract or generate colour palettes from various sources.

## Available plugins

| Plugin | Description | Use case |
|--------|-------------|----------|
| [image](/docs/plugins/input/image) | Extract from images | Local wallpapers, photos |
| [google-genai](/docs/plugins/input/google-genai) | AI generation | Create unique themes |
| [openrouter](/docs/plugins/input/openrouter) | Multi-model AI | Alternative AI generation |
| [remote-json](/docs/plugins/input/remote-json) | JSON URL fetching | Theme repositories, Catppuccin |
| [remote-css](/docs/plugins/input/remote-css) | CSS parsing | CSS variable themes |
| [paletty](./paletty.md) | paletty.dev fetcher (external) | Curated terminal palettes |
| [file](/docs/plugins/input/file) | Manual specification | Custom colour lists |
| [markdown](/docs/plugins/input/markdown) | Saved themes | Restore exported themes |

## Usage

Specify input with `-i` flag:

```bash
tinct generate -i image -p ~/wallpaper.jpg -o all
tinct generate -i remote-json --remote-json.url "..." -o all
```

## Common interface

All input plugins:

1. Accept configuration via flags
2. Validate inputs before processing
3. Return a categorised colour palette
4. Optionally provide theme hints (dark/light)
5. Optionally provide wallpaper path

## Optional features

### Theme hinting

Some plugins can suggest a theme type:

| Plugin | Detection method |
|--------|------------------|
| image | Luminance analysis |
| google-genai | Prompt inference |
| remote-json | Metadata if available |
| file | Based on background colour |

### Wallpaper provision

Some plugins provide a wallpaper for output plugins:

| Plugin | Wallpaper source |
|--------|------------------|
| image | The input image |
| google-genai | Generated image |
| openrouter | Generated image |
| markdown | Embedded wallpaper |

Two template fields are available:

| Field | Description |
|-------|-------------|
| `.WallpaperPath` | Canonical path (resolved relative paths, cached URLs) |
| `.WallpaperRawPath` | Literal user input (original path/URL) |

Output plugins like `hyprpaper` use `.WallpaperPath` to set the wallpaper automatically.

See [Template functions: Context fields](/docs/templating/functions#context-fields) for details.

## Choosing an input

| Scenario | Recommended input |
|----------|-------------------|
| Match your wallpaper | `image` |
| Use a known theme | `remote-json`, `paletty`, or `file` |
| Want something unique | `google-genai` or `openrouter` |
| Restore saved theme | `markdown` |
| Quick custom colours | `file` |
