---
sidebar_position: 2
---

# Template functions

Reference for all tinct template functions.

## Colour access

### `get <palette> <role>`

Get a colour by role name. Panics if role doesn't exist.

```go
{{ get . "background" | hex }}      # #1e1e2e
{{ get . "accent1" | rgb }}         # rgb(137, 180, 250)
```

### `getSafe <palette> <role>`

Safely get a colour. Returns error if role doesn't exist.

```go
{{ $colour, $err := getSafe . "customRole" }}
{{ if $err }}
# Role not found
{{ else }}
colour {{ $colour | hex }}
{{ end }}
```

### `has <palette> <role>`

Check if a role exists.

```go
{{ if has . "border" }}
border {{ get . "border" | hex }}
{{ else }}
border {{ get . "outline" | hex }}
{{ end }}
```

### `getByIndex <palette> <index>`

Get colour by index in the AllColors array (sorted by luminance).

```go
{{ $colour, _ := getByIndex . 0 }}
color0 {{ $colour | hex }}
```

### `ansi <palette> <colorName>`

Find closest palette colour to an ANSI colour name.

Supported names:
- Standard: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`
- Bright: `brightblack`, `brightred`, `brightgreen`, `brightyellow`, `brightblue`, `brightmagenta`, `brightcyan`, `brightwhite`
- Aliases: `color0`-`color15`, `gray`, `grey`, `purple`, etc.
- Extended: `orange`, `pink`, `brown`, `lime`, `navy`, `teal`, `maroon`, `olive`, `violet`, `indigo`

```go
color0 {{ ansi . "black" | hex }}
color1 {{ ansi . "red" | hex }}
color2 {{ ansi . "green" | hex }}
```

### `ansiSafe <palette> <colorName>`

Safely find ANSI colour. Returns error if name not recognized.

```go
{{ $colour, $err := ansiSafe . "customName" }}
```

## Format conversion

All format functions take a ColorValue and return a string.

### `hex <colour>`

Returns `#RRGGBB` format (no alpha).

```go
{{ get . "background" | hex }}
# Output: #1e1e2e
```

### `hexAlpha <colour>`

Returns `#RRGGBBAA` format (with alpha).

```go
{{ get . "scrim" | hexAlpha }}
# Output: #00000052
```

### `hexNoHash <colour>`

Returns `RRGGBB` format (no # prefix).

```go
{{ get . "accent1" | hexNoHash }}
# Output: 89b4fa
```

### `argb <colour>`

Returns `#AARRGGBB` format (alpha-first, for Qt).

```go
{{ get . "background" | argb }}
# Output: #ff1e1e2e
```

### `rgb <colour>`

Returns CSS `rgb(r,g,b)` format.

```go
{{ get . "background" | rgb }}
# Output: rgb(30, 30, 46)
```

### `rgba <colour>`

Returns CSS `rgba(r,g,b,a)` format.

```go
{{ get . "scrim" | rgba }}
# Output: rgba(0, 0, 0, 0.32)
```

### `rgbDecimal <colour>`

Returns `r,g,b` decimal format (for Hyprland).

```go
{{ get . "accent1" | rgbDecimal }}
# Output: 137,180,250
```

### `rgbaDecimal <colour>`

Returns `r,g,b,a` decimal format.

```go
{{ get . "background" | withAlpha 0.93 | rgbaDecimal }}
# Output: 30,30,46,0.93
```

### `rgbSpaces <colour>`

Returns `r g b` space-separated format (for Zellij KDL).

```go
{{ get . "accent1" | rgbSpaces }}
# Output: 137 180 250
```

## Alpha manipulation

### `withAlpha <colour> <alpha>`

Returns a new colour with specified alpha (0.0-1.0).

```go
{{ get . "background" | withAlpha 0.85 | rgba }}
# Output: rgba(30, 30, 46, 0.85)

{{ get . "accent1" | withAlpha 0.5 | hexAlpha }}
# Output: #89b4fa80
```

## Metadata

### `role <colour>`

Get the role name of a colour.

```go
{{ $colour := get . "background" }}
{{ role $colour }}
# Output: background
```

### `index <colour>`

Get the index of a colour in AllColors.

```go
{{ $colour := get . "accent1" }}
{{ index $colour }}
# Output: 4
```

### `themeType <palette>`

Get theme type ("dark" or "light").

```go
{{ themeType . }}
# Output: dark
```

### `count <palette>`

Get total number of colours.

```go
{{ count . }}
# Output: 49
```

## Iteration

### `allRoles <palette>`

Get all role names.

```go
{{ range allRoles . }}
# {{ . }}: {{ get $ . | hex }}
{{ end }}
```

### `allColors <palette>`

Get all colours sorted by luminance.

```go
{{ range $i, $colour := allColors . }}
colour{{ $i }} {{ $colour | hex }}
{{ end }}
```

## Context fields

Templates receive a `ThemeData` struct:

```go
{{- if .WallpaperPath }}
wallpaper {{ .WallpaperPath }}
{{- end }}

{{- if .ThemeName }}
theme {{ .ThemeName }}
{{- end }}
```

## See also

- [Colour access](/docs/templating/color-access)
- [Format conversion](/docs/templating/format-conversion)
