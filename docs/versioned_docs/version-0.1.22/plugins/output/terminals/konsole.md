---
sidebar_position: 4
---

# konsole

Generate colour schemes for [KDE Konsole](https://konsole.kde.org/) terminal.

## Description

Konsole is KDE's terminal emulator. The plugin generates a `.colorscheme` file compatible with Konsole's colour scheme format.

## Output path

```
~/.local/share/konsole/Tinct.colorscheme
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o konsole
```

## Configuration

After generation:

1. Open Konsole
2. Settings > Edit Current Profile
3. Appearance tab
4. Select "Tinct" from the colour scheme dropdown

## Generated format

Konsole uses an INI-style format with colour values as RGB triplets:

```ini
[Background]
Color=30,30,46

[BackgroundFaint]
Color=30,30,46

[BackgroundIntense]
Color=30,30,46

[Color0]
Color=69,71,90

[Color0Faint]
Color=69,71,90

[Color0Intense]
Color=88,91,112

# ... continues for colors 1-7

[Foreground]
Color=205,214,244

[ForegroundFaint]
Color=166,173,200

[ForegroundIntense]
Color=205,214,244

[General]
Description=Tinct Theme
Opacity=1
Wallpaper=
```

## See also

- [Konsole documentation](https://docs.kde.org/stable5/en/konsole/konsole/)
- [kde-plasma](/docs/plugins/output/desktop/kde-plasma)
