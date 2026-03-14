---
sidebar_position: 7
---

# file

Manually specify colours and role assignments.

## Description

The `file` plugin accepts explicit colour specifications with optional role hints. Use it for:

- Known colour palettes (Dracula, Nord, etc.)
- Custom colour schemes
- Precise role assignment

## Usage

```bash
tinct generate -i file --file.colors "<colors>" [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--file.colors` | | Comma-separated hex colours (required) |
| `--file.hints` | | Role assignments: `role=index,role2=index2` |
| `--file.path` | | Load colours from JSON file |

## Examples

### Simple colour list

```bash
tinct generate -i file \
  --file.colors "#1e1e2e,#cdd6f4,#89b4fa,#f5c2e7" \
  -o all
```

### With role hints

```bash
tinct generate -i file \
  --file.colors "#1e1e2e,#cdd6f4,#89b4fa,#f38ba8,#a6e3a1,#f9e2af" \
  --file.hints "background=0,foreground=1,accent1=2,danger=3,success=4,warning=5" \
  -o all
```

### Popular themes

#### Dracula

```bash
tinct generate -i file \
  --file.colors "#282a36,#f8f8f2,#bd93f9,#ff79c6,#50fa7b,#f1fa8c,#8be9fd,#ff5555" \
  --file.hints "background=0,foreground=1,accent1=2,accent2=3,success=4,warning=5,info=6,danger=7" \
  -o all -t dark
```

#### Nord

```bash
tinct generate -i file \
  --file.colors "#2e3440,#eceff4,#88c0d0,#81a1c1,#a3be8c,#ebcb8b,#5e81ac,#bf616a" \
  --file.hints "background=0,foreground=1,accent1=2,accent2=3,success=4,warning=5,info=6,danger=7" \
  -o all -t dark
```

#### Gruvbox Dark

```bash
tinct generate -i file \
  --file.colors "#282828,#ebdbb2,#458588,#b16286,#98971a,#d79921,#689d6a,#cc241d" \
  --file.hints "background=0,foreground=1,info=2,accent2=3,success=4,warning=5,accent1=6,danger=7" \
  -o all -t dark
```

#### Solarized Dark

```bash
tinct generate -i file \
  --file.colors "#002b36,#839496,#268bd2,#d33682,#859900,#b58900,#2aa198,#dc322f" \
  --file.hints "background=0,foreground=1,accent1=2,accent2=3,success=4,warning=5,info=6,danger=7" \
  -o all -t dark
```

### From JSON file

Create `colors.json`:

```json
{
  "colors": ["#1e1e2e", "#cdd6f4", "#89b4fa"],
  "hints": {
    "background": 0,
    "foreground": 1,
    "accent1": 2
  }
}
```

Load it:

```bash
tinct generate -i file --file.path colors.json -o all
```

## Role hints

### Format

```
role=index,role2=index2,...
```

### Available roles

Core roles:
- `background`, `foreground`
- `backgroundMuted`, `foregroundMuted`

Accents:
- `accent1`, `accent2`, `accent3`, `accent4`

Semantic:
- `danger`, `warning`, `success`, `info`

### Example

With colours `#000,#fff,#f00,#0f0,#00f`:

```
--file.hints "background=0,foreground=1,danger=2,success=3,accent1=4"
```

## Auto-categorisation

Colours without hints are auto-assigned:

1. First colour without hint → background
2. Second → foreground
3. Remaining → accents by vibrancy
4. Semantic colours generated if not specified

## Theme type

Specify theme type explicitly:

```bash
# Force dark
tinct generate -i file --file.colors "..." -o all -t dark

# Force light
tinct generate -i file --file.colors "..." -o all -t light
```

## See also

- [remote-json](/docs/plugins/input/remote-json) - Fetch from URL
- [Colour roles](/docs/concepts/color-roles) - All available roles
