"""Real desktop application fixtures sharing the sandbox process lifecycle."""
import configparser
import subprocess
import json
import os
from pathlib import Path
import re
import time
from PIL import Image, ImageDraw
from visual import contrast, missing_text

PALETTES = {
    "dark": {"background": "1e1e2e", "foreground": "cdd6f4", "accent1": "89b4fa", "danger": "f38ba8"},
    "light": {"background": "eff1f5", "foreground": "4c4f69", "accent1": "1e66f5", "danger": "d20f39"},
}


def rgb(value):
    return tuple(bytes.fromhex(value.lstrip("#")[:6]))


def compositor(c):
    proc = c.start("sway", "-c", "/opt/sandbox/sway.conf", log="sway.log")
    runtime = Path(os.environ["XDG_RUNTIME_DIR"])
    c.wait_until(lambda: any(p.is_socket() for p in runtime.glob("wayland-*")) and
                 any(p.is_socket() for p in runtime.glob("sway-ipc.*.sock")), "Wayland and Sway IPC sockets")
    os.environ["WAYLAND_DISPLAY"] = next(p.name for p in runtime.glob("wayland-*") if p.is_socket())
    os.environ["SWAYSOCK"] = str(next(runtime.glob("sway-ipc.*.sock")))
    c.check(proc.poll() is None, "headless compositor running")


def generate(c, app, theme, relative):
    argv = ["tinct", "generate", "-i", "file", "-o", app, "-t", theme, "--force"]
    # A full input palette exercises roles required by terminal/launcher templates.
    palette = dict(PALETTES[theme], accent2="c678dd", accent3="56b6c2", accent4="e5c07b",
                   success="98c379", warning="e5c07b", info="61afef")
    for key, value in palette.items():
        argv += ["--colour", key + "=#" + value]
    c.run(*argv, timeout=5)
    path = Path.home() / ".config" / relative
    c.check(path.is_file() and path.stat().st_size > 100, theme + " generated " + relative)
    (c.RESULTS / (theme + "-" + path.name)).write_bytes(path.read_bytes())
    return path


def measure(image, regions):
    image = image.convert("RGB")
    records = []
    overlay = image.copy()
    draw = ImageDraw.Draw(overlay)
    for name, box, colour, minimum, *options in regions:
        tolerance = options[0] if options else 0
        crop = image.crop(box)
        counts = {colour: count for count, colour in crop.getcolors(crop.width * crop.height)}
        value = sum(count for pixel, count in counts.items()
                    if max(abs(a - b) for a, b in zip(pixel, colour)) <= tolerance) / (crop.width * crop.height)
        passed = value >= minimum
        records.append(dict(name=name, region=box, value=value, minimum=minimum, tolerance=tolerance, passed=passed))
        draw.rectangle(box, outline="lime" if passed else "red", width=2)
    return records, overlay


def capture(c, name, regions, crop_box, expected, text_colours, ocr_boxes=None):
    time.sleep(1)
    path = c.RESULTS / (name + ".png")
    c.run("grim", str(path))
    with Image.open(path) as image:
        c.check(image.size == (1280, 900), name + " dimensions")
        records, overlay = measure(image, regions)
        overlay.save(c.RESULTS / (name + "-regions.png"))
        ocr_paths = []
        for index, box in enumerate(ocr_boxes or [crop_box]):
            crop = image.crop(box).convert("RGB")
            ocr_path = c.RESULTS / (name + f"-ocr-{index}.png")
            crop.resize((crop.width * 3, crop.height * 3)).save(ocr_path)
            ocr_paths.append(ocr_path)
    for foreground, background in text_colours:
        ratio = contrast(foreground, background)
        records.append(dict(name="text contrast", value=ratio, minimum=4.5, passed=ratio >= 4.5))
    report = {"metrics": records}
    report_path = c.RESULTS / (name + "-metrics.json")
    report_path.write_text(json.dumps(report, indent=2))
    text = "\n".join(c.run("tesseract", str(path), "stdout", "-l", "eng", "--psm", "6") for path in ocr_paths)
    report["ocr"] = dict(text=text, expected=expected, missing=missing_text(text, expected))
    report["passed"] = all(r["passed"] for r in records) and not report["ocr"]["missing"]
    report_path.write_text(json.dumps(report, indent=2))
    for record in records:
        c.check(record["passed"], name + " " + record["name"])
    c.check(not report["ocr"]["missing"], name + " visible text " + str(report["ocr"]["missing"]))


def waybar(c):
    compositor(c)
    config = Path.home() / ".config/waybar"
    (config / "themes").mkdir(parents=True)
    fixture = {"layer": "top", "position": "top", "height": 40,
               "modules-left": ["custom/title"], "modules-center": ["custom/media"],
               "modules-right": ["clock"], "custom/title": {"format": "Tinct status"},
               "custom/media": {"format": "Music ready"}, "clock": {"format": "Clock ready"}}
    (config / "config.json").write_text(json.dumps(fixture))
    # Exercise the generated example stylesheet, not a replacement theme.
    style = '@import "tinct-style.css";\n* { font-family: "DejaVu Sans"; font-size: 16px; }\n'
    (config / "style.css").write_text(style)
    daemon = None
    for theme in ("dark", "light"):
        path = generate(c, "waybar", theme, "waybar/themes/tinct.css")
        colours = {name: rgb(value) for name, value in re.findall(r"@define-color ([\w-]+) (#[0-9a-fA-F]{6});", path.read_text())}
        for role in ("background", "foreground"):
            c.check(colours[role] == rgb(PALETTES[theme][role]), theme + " generated requested " + role)
        if daemon is None:
            daemon = c.start("waybar", "-c", str(config / "config.json"), "-s", str(config / "style.css"), log="waybar.log")
        c.check(daemon.poll() is None, theme + " same Waybar process alive")
        capture(c, theme + "-waybar", [
            ("bar background", (200, 2, 400, 38), colours["background"], .95),
            ("title foreground", (0, 0, 180, 40), colours["foreground"], .005),
            ("module surface", (580, 2, 700, 38), colours["surface"], .30),
            ("module text", (580, 2, 700, 38), colours["onSurface"], .005),
            ("desktop preserved", (0, 60, 1280, 900), (48, 52, 59), .99),
        ], (0, 0, 1280, 50), ["Tinct status", "Music ready", "Clock ready"],
        [(colours["foreground"], colours["background"]), (colours["onSurface"], colours["surface"])])
    c.check((config / "style.css").read_text() == style, "Waybar user stylesheet preserved")
    log = (c.RESULTS / "waybar.log").read_text().lower()
    c.check(not any(marker in log for marker in ("[error]", "theme parsing error")), "Waybar has no runtime/CSS errors")


def fuzzel(c):
    compositor(c)
    config = Path.home() / ".config/fuzzel"
    config.mkdir(parents=True, exist_ok=True)
    entries = c.RESULTS / "entries.txt"
    entries.write_text("Alpha editor\nBeta browser\nGamma terminal\n")
    for theme in ("dark", "light"):
        path = generate(c, "fuzzel", theme, "fuzzel/themes/tinct.ini")
        ini = configparser.ConfigParser()
        ini.read(path)
        colours = {key: rgb(value) for key, value in ini["colors"].items()}
        for key, role in (("background", "background"), ("text", "foreground"), ("selection", "accent1")):
            c.check(colours[key] == rgb(PALETTES[theme][role]), theme + " requested " + key)
        # Preserve generated opacity; compare composited pixels against the known desktop.
        alpha = int(ini["colors"]["background"][6:8], 16) / 255
        background = tuple(int(a * alpha) + round(b * (1 - alpha)) for a, b in zip(colours["background"], (48, 52, 59)))
        main = config / "fuzzel.ini"
        content = f"include={path}\nfont=DejaVu Sans Mono:size=14\ndpi-aware=no\nanchor=top-left\nx-margin=40\ny-margin=100\nwidth=40\nlines=3\nline-height=32\nhorizontal-pad=16\nvertical-pad=8\n[border]\nwidth=2\nradius=0\n"
        main.write_text(content)
        c.run("fuzzel", "--config", str(main), "--check-config")
        result = c.RESULTS / (theme + "-selection.txt")
        with entries.open() as source, result.open("w") as out, (c.RESULTS / (theme + "-fuzzel.log")).open("w") as log:
            proc = subprocess.Popen(["fuzzel", "--config", str(main), "--dmenu", "--no-sort", "--prompt", "Find: "], stdin=source, stdout=out, stderr=log)
        c.PROCESSES.append(proc)
        def shot(state, selected_y, other_y, expected):
            regions = [
                ("launcher background", (180, 110, 450, 135), background, .80, 1),
                ("selected background", (60, selected_y, 450, selected_y + 40), colours["selection"], .60),
                ("selected text", (60, selected_y, 450, selected_y + 40), colours["selection-text"], .002),
                ("unselected text", (60, other_y, 450, other_y + 40), colours["text"], .002),
                ("desktop preserved", (700, 0, 1280, 900), (48, 52, 59), .99),
            ]
            capture(c, theme + "-fuzzel-" + state, regions, (40, 100, 650, 300), expected,
                    [(colours["selection-text"], colours["selection"]), (colours["text"], background)],
                    ocr_boxes=[(50, y, 505, y + 40) for y in (110, 153, 196, 239)])
            c.check(proc.poll() is None, state + " Fuzzel running")
        shot("initial", 153, 196, ["Find", "Alpha editor", "Beta browser", "Gamma terminal"])
        c.run("wtype", "-k", "Down")
        shot("navigated", 196, 153, ["Alpha editor", "Beta browser", "Gamma terminal"])
        c.run("wtype", "Beta")
        time.sleep(.5)
        c.run("wtype", "-k", "Return")
        c.check(proc.wait(timeout=5) == 0, theme + " Fuzzel selection accepted")
        c.check(result.read_text().strip() == "Beta browser", theme + " keyboard search selects expected entry")
        c.check(main.read_text() == content, "Fuzzel user config preserved")
