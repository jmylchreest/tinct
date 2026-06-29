---
sidebar_position: 1
---

# Input plugins

Input plugins extract or generate colour palettes from various sources.

## Available plugins

{/* AUTO-GENERATED PLUGIN TABLE — DO NOT EDIT BETWEEN THE MARKERS.
    Regenerated from in-tree READMEs by docs/scripts/sync-plugin-readmes.mjs.
    To change the table format, edit the script's `renderInputTable` function. */}
{/* BEGIN AUTO-PLUGIN-TABLE */}

| Plugin | Description | Source type | Requires |
|--------|-------------|-------------|----------|
| [image](./image.md) | Extract colours from images (files, directories, URLs) with optional ambient lighting | image |  |
| [file](./file.md) | Load palette from file or colour overrides | file |  |
| [markdown](./markdown.md) | Load theme from tinct markdown theme file | file |  |
| [remote-json](./remote-json.md) | Fetch colours from remote JSON with JSONPath queries | url | network |
| [remote-css](./remote-css.md) | Fetch colours from remote CSS variables | url | network |
| [google-genai](./google-genai.md) | Generate images with Google Imagen and extract colours | api | network; creds: GOOGLE_API_KEY |
| [openrouter](./openrouter.md) | Generate images with OpenRouter.ai models and extract colours | api | network; creds: OPENROUTER_API_KEY |
| [minimax](./minimax.md) | Generate images with MiniMax (Hailuo) and extract colours | api | network; creds: MINIMAX_API_KEY |
| [paletty](./paletty.md) | Fetch curated terminal palettes from paletty.dev | api | network |
| [random](./random.md) | Generate random colour palettes with configurable seed and colour count | generated |  |

{/* END AUTO-PLUGIN-TABLE */}

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

See [Template functions: Context fields](../../templating/functions.md#context-fields) for details.

## Choosing an input

| Scenario | Recommended input |
|----------|-------------------|
| Match your wallpaper | `image` |
| Want something unique | `google-genai` or `openrouter` |
| Restore saved theme | `markdown` |
| Quick custom colours | `file` |
