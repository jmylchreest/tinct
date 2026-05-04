---
sidebar_position: 2
---

# Material Design 3

Tinct follows Material Design 3 (M3) guidelines for colour categorisation, ensuring consistent and accessible theming across applications.

## What is Material Design 3?

Material Design 3 is Google's latest design system, introducing:

- **Dynamic colour**: Colours derived from user content
- **Semantic roles**: Colours defined by their purpose, not just appearance
- **Accessibility**: Built-in contrast requirements

Tinct adopts these principles to generate harmonious, accessible colour schemes.

## Semantic colour roles

Instead of arbitrary colour names, M3 uses semantic roles:

| Role | Purpose |
|------|---------|
| `background` | Main app background |
| `surface` | Card and sheet backgrounds |
| `primary` | Main brand/accent colour |
| `danger` | Error states and destructive actions |
| `warning` | Warning and caution states |
| `success` | Success and confirmation states |

Each role has defined relationships to other roles, ensuring consistency.

## WCAG compliance

Tinct implements WCAG 2.1 contrast checking:

- **AA standard**: Minimum 4.5:1 contrast for normal text
- **AAA standard**: Minimum 7:1 contrast for enhanced accessibility

When assigning colours to roles, tinct verifies:

1. Foreground/background pairs meet contrast requirements
2. "On-colour" variants (text on coloured backgrounds) are readable
3. Surface colours maintain hierarchy through contrast

## Surface elevation

M3 defines surface containers at different elevation levels:

| Role | Elevation | Purpose |
|------|-----------|---------|
| `surfaceContainerLowest` | 0 | Base surface |
| `surfaceContainerLow` | 1 | Slightly raised |
| `surfaceContainer` | 2 | Default cards |
| `surfaceContainerHigh` | 3 | Prominent elements |
| `surfaceContainerHighest` | 4 | Dialogs, menus |

Tinct generates these automatically from the base surface colour.

## On-colours

Every coloured surface needs readable text. M3 defines "on-colour" pairs:

| Surface | Text colour |
|---------|-------------|
| `background` | `foreground` |
| `surface` | `onSurface` |
| `accent1` | `onAccent1` |
| `danger` | `onDanger` |

Tinct calculates these to ensure readability.

## Inverse colours

For tooltips, snackbars, and overlays that need to stand out:

| Role | Purpose |
|------|---------|
| `inverseSurface` | Inverted surface colour |
| `inverseOnSurface` | Text on inverse surface |
| `inversePrimary` | Accent on inverse surface |

## Special effects

Two colours include transparency for specific effects:

| Role | Alpha | Purpose |
|------|-------|---------|
| `scrim` | 32% | Modal backdrop overlay |
| `shadow` | 15% | Drop shadow colour |

## Applying M3 to applications

Not all applications support the full M3 colour system. Tinct maps roles to application-specific values:

### GTK4/Libadwaita

Libadwaita natively supports M3-style theming with CSS variables.

### KDE Plasma

Plasma uses a different colour model. Tinct maps M3 roles to Plasma's:

- Button, View, Window sections
- Selection and tooltip colours
- Window manager decorations

### Terminals

Terminal emulators use ANSI colour palettes (16 colours). Tinct maps semantic roles to the closest ANSI colour.

## Customisation

Override automatic categorisation with role hints:

```bash
tinct generate -i file \
  --file.colors "#1e1e2e,#cdd6f4,#89b4fa" \
  --file.hints "background=0,foreground=1,accent1=2" \
  -o all
```

## Next steps

- [Colour roles](/docs/concepts/color-roles) - Full list of all roles
- [Theme detection](/docs/concepts/theme-detection) - Light/dark detection
- [Templating](/docs/templating) - Access colours in templates
