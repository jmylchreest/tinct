---
sidebar_position: 3
---

# Colour roles

Tinct defines 49+ semantic colour roles, each with a specific purpose. This page documents all available roles.

## Core colours (4)

The foundation of every theme:

| Role | Description |
|------|-------------|
| `background` | Base background colour |
| `backgroundMuted` | Muted/inactive background |
| `foreground` | Primary text colour |
| `foregroundMuted` | Secondary/dimmed text |

## Accent colours (8)

Vibrant colours for emphasis and branding:

| Role | Description |
|------|-------------|
| `accent1` | Primary accent |
| `accent1Muted` | Muted variant |
| `accent2` | Secondary accent |
| `accent2Muted` | Muted variant |
| `accent3` | Tertiary accent |
| `accent3Muted` | Muted variant |
| `accent4` | Quaternary accent |
| `accent4Muted` | Muted variant |

## Semantic colours (5)

Colours with specific meaning:

| Role | Purpose | Typical colour |
|------|---------|----------------|
| `danger` | Error, destructive actions | Red |
| `warning` | Warnings, caution | Orange/Yellow |
| `success` | Success, confirmation | Green |
| `info` | Information | Blue |
| `notification` | Notifications | Purple |

## Surface colours - Priority 1 (4)

Essential UI surface colours:

| Role | Description |
|------|-------------|
| `surface` | Card/sheet backgrounds |
| `onSurface` | Text on surface |
| `outline` | Borders/dividers |
| `border` | Primary borders |

## Surface variants - Priority 2 (4)

Extended surface options:

| Role | Description |
|------|-------------|
| `surfaceVariant` | Alternate surface |
| `onSurfaceVariant` | Text on variant |
| `borderMuted` | Inactive borders |
| `outlineVariant` | Secondary outlines |

## On-colours for accents (4)

Text colours for accent backgrounds:

| Role | Description |
|------|-------------|
| `onAccent1` | Text on accent1 |
| `onAccent2` | Text on accent2 |
| `onAccent3` | Text on accent3 |
| `onAccent4` | Text on accent4 |

## On-colours for semantics (4)

Text colours for semantic backgrounds:

| Role | Description |
|------|-------------|
| `onDanger` | Text on danger |
| `onWarning` | Text on warning |
| `onSuccess` | Text on success |
| `onInfo` | Text on info |

## Inverse colours (3)

For overlays and inverted contexts:

| Role | Description |
|------|-------------|
| `inverseSurface` | Tooltip/inverse surface |
| `inverseOnSurface` | Text on inverse |
| `inversePrimary` | Accent on inverse |

## Overlay colours (2)

Transparency effects:

| Role | Alpha | Description |
|------|-------|-------------|
| `scrim` | 32% | Modal backdrop |
| `shadow` | 15% | Elevation shadow |

## Container elevation (5)

Surface hierarchy through elevation:

| Role | Level | Description |
|------|-------|-------------|
| `surfaceContainerLowest` | 0 | Base level |
| `surfaceContainerLow` | 1 | Slightly raised |
| `surfaceContainer` | 2 | Default cards |
| `surfaceContainerHigh` | 3 | Prominent |
| `surfaceContainerHighest` | 4 | Dialogs, menus |

## Positional colours (30+)

For ambient lighting and LED strips:

### Core 8 positions

```
positionTopLeft      positionTop      positionTopRight
positionLeft                          positionRight
positionBottomLeft   positionBottom   positionBottomRight
```

### Extended 12+ positions

Additional positions for finer-grained control:

```
positionTopLeftInner    positionTopCenter    positionTopRightInner
positionCenterLeft      positionCenter       positionCenterRight
positionBottomLeftInner positionBottomCenter positionBottomRightInner
```

### 16+ ultra-extended positions

Full grid including additional edge midpoints.

## Using roles in templates

Access roles in output plugin templates:

```go
# Core colours
{{ get . "background" | hex }}
{{ get . "foreground" | hex }}

# Semantic colours
{{ get . "danger" | hex }}
{{ get . "success" | hex }}

# With alpha
{{ get . "scrim" | rgba }}
```

## Role priority

When extracting from images, roles are assigned in priority order:

1. **Priority 0** (Core): background, foreground
2. **Priority 1** (Accents): accent1-4
3. **Priority 2** (Semantic): danger, warning, success, info
4. **Priority 3** (Surface): Generated from background
5. **Priority 4** (On-colours): Calculated for contrast
6. **Positional**: Extracted separately with ambient mode

## Manual role assignment

Override automatic assignment:

```bash
tinct generate -i file \
  --file.colors "#1e1e2e,#cdd6f4,#89b4fa,#f38ba8" \
  --file.hints "background=0,foreground=1,accent1=2,danger=3" \
  -o all
```

## Next steps

- [Theme detection](./theme-detection.md) - Light/dark detection
- [Templating](../templating/index.md) - Use roles in custom templates
- [Format conversion](../templating/format-conversion.md) - Output formats
