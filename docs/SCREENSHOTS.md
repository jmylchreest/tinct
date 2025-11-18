# Adding Screenshots to Tinct Documentation

This guide explains how to add and organize screenshots for the tinct project.

## Directory Structure

Screenshots should be organized in the `.github/` directory:

```
tinct/
├── .github/
│   └── screenshots/
│       ├── quick-start/
│       │   ├── extract-preview.png
│       │   ├── generate-all.png
│       │   └── themed-desktop.png
│       ├── plugins/
│       │   ├── hyprland-themed.png
│       │   ├── kitty-themed.png
│       │   ├── kde-plasma-themed.png
│       │   └── qt-apps-themed.png
│       ├── features/
│       │   ├── color-extraction.png
│       │   ├── ambient-lighting.png
│       │   └── theme-preview.png
│       └── banner.png
├── README.md
└── docs/
    └── SCREENSHOTS.md (this file)
```

## GitHub Best Practices

### 1. Use Relative Paths

Always use relative paths from the repository root:

```markdown
![Tinct Banner](.github/screenshots/banner.png)
![Extract Preview](.github/screenshots/quick-start/extract-preview.png)
```

### 2. Include Alt Text

Always provide descriptive alt text for accessibility:

```markdown
![Screenshot showing tinct extracting colors from a mountain landscape wallpaper](.github/screenshots/features/color-extraction.png)
```

### 3. Optimize Image Sizes

- **Banner/Hero images:** 1200px width recommended
- **Feature screenshots:** 800-1000px width
- **UI examples:** 600-800px width
- **Format:** PNG for UI screenshots, JPEG for photos
- **Compression:** Use tools like `pngquant` or `optipng` to reduce file size

```bash
# Optimize PNG screenshots
pngquant --quality=80-95 screenshot.png -o screenshot-optimized.png
optipng -o7 screenshot-optimized.png

# Or use ImageMagick for batch processing
mogrify -resize 1000x -quality 90 *.png
```

### 4. Dark Mode Considerations

For screenshots that may appear in dark mode:

```markdown
<!-- Use HTML img tag with theme-aware styling -->
<picture>
  <source media="(prefers-color-scheme: dark)" srcset=".github/screenshots/screenshot-dark.png">
  <source media="(prefers-color-scheme: light)" srcset=".github/screenshots/screenshot-light.png">
  <img alt="Tinct themed desktop" src=".github/screenshots/screenshot-light.png">
</picture>
```

Or use GitHub's theme-aware syntax:

```markdown
![Screenshot](.github/screenshots/screenshot.png#gh-light-mode-only)
![Screenshot](.github/screenshots/screenshot-dark.png#gh-dark-mode-only)
```

## Recommended Screenshots

### Essential Screenshots

1. **Banner/Hero Image** (`.github/screenshots/banner.png`)
   - Should show tinct's primary use case
   - Consider: Before/after themed desktop comparison
   - Size: 1200x630px (GitHub social preview size)

2. **Color Extraction** (`.github/screenshots/quick-start/extract-preview.png`)
   - `tinct extract --preview --categorise wallpaper.jpg`
   - Shows the terminal output with color palette
   - Demonstrates role categorization

3. **Themed Desktop** (`.github/screenshots/quick-start/themed-desktop.png`)
   - Full desktop screenshot showing multiple themed applications
   - Should include: terminal, launcher, bar, notifications
   - Demonstrates the unified theming capability

### Plugin Screenshots

4. **Hyprland Setup** (`.github/screenshots/plugins/hyprland-themed.png`)
   - Hyprland desktop with themed waybar, launcher, and terminal
   
5. **KDE Plasma Setup** (`.github/screenshots/plugins/kde-plasma-themed.png`)
   - KDE Plasma desktop showing themed panels and applications
   
6. **Qt Applications** (`.github/screenshots/plugins/qt-apps-themed.png`)
   - Qt-based applications (e.g., Zapzap, Dolphin) with tinct theme applied

7. **Terminal Comparison** (`.github/screenshots/plugins/terminals.png`)
   - Side-by-side of Kitty, Alacritty, Ghostty with same theme

### Feature Screenshots

8. **Ambient Lighting** (`.github/screenshots/features/ambient-lighting.png`)
   - LED strip synchronized with wallpaper colors
   - Show physical setup with monitor and LED bias lighting

9. **Theme Preview** (`.github/screenshots/features/theme-preview.png`)
   - Output of categorized palette with role assignments
   - Helpful for understanding color roles

10. **AI Generation** (`.github/screenshots/features/ai-generation.png`)
    - Example of Google Gemini-generated image and resulting theme

## Example Usage in README

### Hero Section

```markdown
# Tinct

> An extensible colour palette generator and theme manager

![Tinct Banner - Unified desktop theming from a single wallpaper](.github/screenshots/banner.png)
```

### Quick Start Section

```markdown
## Quick Start

Extract and preview colors from any image:

```bash
tinct extract --preview --categorise wallpaper.jpg
```

![Color extraction preview showing categorized palette](.github/screenshots/quick-start/extract-preview.png)

Apply to your entire desktop environment:

```bash
tinct generate -i image -p wallpaper.jpg -o all
```

![Themed desktop with consistent colors across all applications](.github/screenshots/quick-start/themed-desktop.png)
```

### Features Section

```markdown
## Features

### Unified Theming

Apply consistent color schemes across your entire environment:

<p align="center">
  <img src=".github/screenshots/plugins/hyprland-themed.png" alt="Hyprland themed desktop" width="45%">
  <img src=".github/screenshots/plugins/kde-plasma-themed.png" alt="KDE Plasma themed desktop" width="45%">
</p>

### Ambient LED Synchronization

Extract positional colors for bias lighting:

![LED strip synchronized with wallpaper colors](.github/screenshots/features/ambient-lighting.png)
```

## Creating Screenshots

### Terminal Output

For terminal screenshots, use a terminal with good color reproduction:

```bash
# Use Kitty for crisp terminal screenshots
kitty --config NONE \
  -o font_family="JetBrains Mono" \
  -o font_size=12 \
  -e sh -c 'tinct extract --preview --categorise wallpaper.jpg'
```

### Desktop Screenshots

For desktop screenshots:

```bash
# Wayland (Hyprland, GNOME, KDE Plasma)
grim -g "$(slurp)" screenshot.png

# Or full screen
grim screenshot.png

# X11
scrot screenshot.png
```

### GIFs/Animations

For animated demonstrations:

```bash
# Record terminal with asciinema
asciinema rec tinct-demo.cast

# Convert to GIF (requires agg)
agg tinct-demo.cast tinct-demo.gif

# Or use peek for screen recording
peek
```

## File Naming Conventions

- Use kebab-case: `themed-desktop.png` not `Themed_Desktop.PNG`
- Be descriptive: `hyprland-waybar-themed.png` not `screenshot1.png`
- Include context: `kde-plasma-dark-theme.png` vs `kde-plasma-light-theme.png`

## Commit Guidelines

When adding screenshots:

```bash
git add .github/screenshots/
git commit -m "docs: add screenshots for Qt/KDE plugins

- Add KDE Plasma themed desktop screenshot
- Add Qt applications (Zapzap, Dolphin) themed screenshot
- Optimize all PNGs with pngquant
"
```

## Image Size Guidelines

| Type | Max Width | Max File Size | Format |
|------|-----------|---------------|--------|
| Banner | 1200px | 500KB | PNG/JPEG |
| Desktop screenshots | 1920px | 1MB | PNG |
| Terminal screenshots | 1000px | 300KB | PNG |
| Feature demos | 800px | 500KB | PNG |
| UI elements | 600px | 200KB | PNG |
| Animations | 800px | 5MB | GIF |

## Accessibility

1. **Alt Text:** Always provide meaningful descriptions
2. **Color Contrast:** Ensure text in screenshots is readable
3. **Size:** Images should be viewable at different zoom levels
4. **Captions:** Consider adding captions for complex screenshots

## Example PR Template

When submitting screenshots in a PR:

```markdown
## Screenshots

### Added
- Banner image showing unified desktop theming
- Color extraction preview demonstrating role categorization
- KDE Plasma themed desktop with tinct-generated color scheme

### Technical Details
- All images optimized with pngquant (quality: 80-95)
- Maximum file size: 800KB
- Format: PNG for UI, JPEG for photos
- Responsive sizing included

### Checklist
- [ ] Images optimized for file size
- [ ] Alt text provided for all images
- [ ] Relative paths used
- [ ] Images organized in appropriate subdirectories
- [ ] Dark mode compatibility considered
```

## Tools

Recommended tools for creating and optimizing screenshots:

- **pngquant:** Lossy PNG compression
- **optipng:** Lossless PNG optimization
- **ImageMagick:** Batch processing and resizing
- **peek:** Screen recording for GIFs
- **asciinema:** Terminal recording
- **agg:** Convert asciinema to GIF
- **grim/slurp:** Wayland screenshot tools
- **scrot:** X11 screenshot tool
