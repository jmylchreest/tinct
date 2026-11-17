---
sidebar_position: 3
---

# Colour access

Patterns for accessing colours in templates.

## By role name

The most common way to access colours:

```go
# Core
{{ get . "background" | hex }}
{{ get . "foreground" | hex }}

# Accents
{{ get . "accent1" | hex }}
{{ get . "accent2" | hex }}

# Semantic
{{ get . "danger" | hex }}
{{ get . "warning" | hex }}
{{ get . "success" | hex }}
{{ get . "info" | hex }}

# Surface
{{ get . "surface" | hex }}
{{ get . "onSurface" | hex }}
```

## With fallbacks

Check if a role exists before using:

```go
{{ if has . "border" }}
border {{ get . "border" | hex }}
{{ else }}
border {{ get . "outline" | hex }}
{{ end }}
```

## By index

Access colours by luminance-sorted index:

```go
{{- range $i := seq 0 15 }}
{{- $colour, _ := getByIndex . $i }}
colour{{ $i }} {{ $colour | hex }}
{{- end }}
```

## ANSI terminal colours

Find closest matches to standard terminal colours:

```go
# Standard 16 colours
{{ range $i, $name := list "black" "red" "green" "yellow" "blue" "magenta" "cyan" "white" }}
colour{{ $i }} {{ ansi $ $name | hex }}
{{ end }}

{{ range $i, $name := list "brightblack" "brightred" "brightgreen" "brightyellow" "brightblue" "brightmagenta" "brightcyan" "brightwhite" }}
colour{{ add $i 8 }} {{ ansi $ $name | hex }}
{{ end }}
```

## Complete role list

### Core (4)

| Role | Description |
|------|-------------|
| `background` | Base background |
| `backgroundMuted` | Inactive background |
| `foreground` | Primary text |
| `foregroundMuted` | Secondary text |

### Accents (8)

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

### Semantic (5)

| Role | Description |
|------|-------------|
| `danger` | Errors, destructive actions |
| `warning` | Warnings, caution |
| `success` | Success, confirmation |
| `info` | Information |
| `notification` | Notifications |

### Surface (8)

| Role | Description |
|------|-------------|
| `surface` | Card backgrounds |
| `onSurface` | Text on surface |
| `surfaceVariant` | Alternate surface |
| `onSurfaceVariant` | Text on variant |
| `outline` | Borders, dividers |
| `outlineVariant` | Secondary borders |
| `border` | Primary borders |
| `borderMuted` | Inactive borders |

### On-colours (8)

| Role | Description |
|------|-------------|
| `onAccent1` | Text on accent1 |
| `onAccent2` | Text on accent2 |
| `onAccent3` | Text on accent3 |
| `onAccent4` | Text on accent4 |
| `onDanger` | Text on danger |
| `onWarning` | Text on warning |
| `onSuccess` | Text on success |
| `onInfo` | Text on info |

### Inverse (3)

| Role | Description |
|------|-------------|
| `inverseSurface` | Tooltip background |
| `inverseOnSurface` | Text on inverse |
| `inversePrimary` | Inverse accent |

### Overlay (2)

| Role | Description |
|------|-------------|
| `scrim` | Modal backdrop (32% opacity) |
| `shadow` | Elevation shadow (15% opacity) |

### Container elevation (5)

| Role | Description |
|------|-------------|
| `surfaceContainerLowest` | Elevation 0 |
| `surfaceContainerLow` | Elevation 1 |
| `surfaceContainer` | Elevation 2 |
| `surfaceContainerHigh` | Elevation 3 |
| `surfaceContainerHighest` | Elevation 4 |

### Positional (30+)

Used for ambient lighting effects:

```go
positionTopLeft
positionTop
positionTopRight
positionRight
positionBottomRight
positionBottom
positionBottomLeft
positionLeft
positionCenter
# ... and more
```

## Using variables

Cache colour lookups for efficiency:

```go
{{- $bg := get . "background" }}
{{- $fg := get . "foreground" }}
{{- $accent := get . "accent1" }}

background {{ $bg | hex }}
foreground {{ $fg | hex }}
accent {{ $accent | hex }}

# Reuse with different formats
selection_bg {{ $accent | withAlpha 0.3 | rgba }}
```

## See also

- [Functions reference](./functions.md)
- [Format conversion](./format-conversion.md)
