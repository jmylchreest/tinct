"""Quantitative checks for the fixed 1280x900 SwayNC test fixture."""
from PIL import ImageDraw


def contrast(first, second):
    def luminance(rgb):
        channels = [v / 255 for v in rgb]
        linear = [v / 12.92 if v <= .04045 else ((v + .055) / 1.055) ** 2.4 for v in channels]
        return sum(v * weight for v, weight in zip(linear, (.2126, .7152, .0722)))
    high, low = sorted((luminance(first), luminance(second)), reverse=True)
    return (high + .05) / (low + .05)


def analyse(image, theme, surface, panel, critical):
    image = image.convert("RGB")
    foreground = (205, 214, 244) if theme == "dark" else (76, 79, 105)
    background = (30, 30, 46) if theme == "dark" else (239, 241, 245)
    danger = (243, 139, 168) if theme == "dark" else (210, 15, 57)
    metrics = []
    overlay = image.copy()
    draw = ImageDraw.Draw(overlay)

    def measure(name, value, minimum, maximum=None, box=None):
        passed = value >= minimum and (maximum is None or value <= maximum)
        metrics.append(dict(name=name, value=value, minimum=minimum, maximum=maximum,
                            passed=passed, region=box))
        if box:
            draw.rectangle(box, outline="lime" if passed else "red", width=2)

    def coverage(name, box, colour, minimum, maximum=None):
        crop = image.crop(box)
        counts = {rgb: count for count, rgb in crop.getcolors(crop.width * crop.height)}
        value = counts.get(colour, 0) / (crop.width * crop.height)
        measure(name, value, minimum, maximum, box)

    measure("width", image.width, 1280, 1280)
    measure("height", image.height, 900, 900)
    # Sample components individually so large unchanged areas cannot hide regressions.
    cards = (810, 145, 1250, 265) if panel else (795, 5, 1265, 120 if critical else 55)
    coverage("notification_surface_fraction", cards, surface, .30, .97)
    coverage("notification_text_fraction", cards, foreground, .002, .20)
    measure("notification_contrast_ratio", contrast(foreground, surface), 4.5)
    coverage("desktop_preserved_fraction", (0, 0, 760, 900), (48, 52, 59), .99)
    if panel:
        coverage("panel_background_fraction", (790, 300, 1270, 880), background, .95)
        coverage("panel_header_text_fraction", (790, 10, 1265, 120), foreground, .002, .20)
        measure("panel_contrast_ratio", contrast(foreground, background), 4.5)
    if critical:
        border = (810, 145, 1250, 200) if panel else (792, 5, 1268, 57)
        coverage("critical_border_fraction", border, danger, .002, .10)
    return {"metrics": metrics, "passed": all(m["passed"] for m in metrics)}, overlay


def missing_text(actual, expected):
    normalised = " ".join(actual.casefold().split())
    return [phrase for phrase in expected if " ".join(phrase.casefold().split()) not in normalised]
