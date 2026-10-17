"""Container-side validators. Every process uses argv and a bounded lifetime."""
import json
import os
import re
from pathlib import Path
import shutil
import subprocess
import sys
import time
import traceback
from PIL import Image
from visual import analyse, missing_text
import desktop

RESULTS = Path("/tmp/results")
PROCESSES = []
CHECKS = []


def run(*argv, timeout=10):
    with (RESULTS / "commands.log").open("a") as log:
        log.write(json.dumps(list(argv)) + "\n")
        log.flush()
        try:
            result = subprocess.run(argv, capture_output=True, text=True, timeout=timeout)
        except subprocess.TimeoutExpired as error:
            for output in (error.stdout, error.stderr):
                if output:
                    log.write(output.decode(errors="replace") if isinstance(output, bytes) else output)
            log.write("COMMAND TIMED OUT\n")
            raise
        log.write(result.stdout + result.stderr)
    result.check_returncode()
    return result.stdout


def start(*argv, log):
    with (RESULTS / log).open("w") as stream:
        proc = subprocess.Popen(argv, stdout=stream, stderr=subprocess.STDOUT)
    PROCESSES.append(proc)
    return proc


def check(condition, description):
    if not condition:
        raise RuntimeError(description)
    CHECKS.append(description)


def wait_until(predicate, description):
    deadline = time.monotonic() + 10
    while time.monotonic() < deadline:
        if predicate():
            return
        time.sleep(.1)
    raise RuntimeError("Timed out: " + description)


def owns_notifications():
    return "boolean true" in run("dbus-send", "--session", "--print-reply",
        "--dest=org.freedesktop.DBus", "/org/freedesktop/DBus",
        "org.freedesktop.DBus.NameHasOwner", "string:org.freedesktop.Notifications")


def generate(theme):
    palettes = {
        "dark": ("#1e1e2e", "#cdd6f4", "#89b4fa", "#f38ba8"),
        "light": ("#eff1f5", "#4c4f69", "#1e66f5", "#d20f39"),
    }
    argv = ["tinct", "generate", "-i", "file", "-o", "swaync", "-t", theme, "--force"]
    for role, colour in zip(("background", "foreground", "accent1", "danger"), palettes[theme]):
        argv += ["--colour", role + "=" + colour]
    run(*argv, timeout=5)
    css = Path.home() / ".config/swaync/tinct.css"
    check(css.is_file() and css.stat().st_size > 100, theme + " stylesheet generated")
    shutil.copyfile(css, RESULTS / (theme + ".css"))
    return tuple(map(int, re.search(r"--noti-bg:\s*(\d+),(\d+),(\d+)", css.read_text()).groups()))


def screenshot(name, theme, surface, panel=False, critical=False):
    time.sleep(1)
    path = RESULTS / (name + ".png")
    run("grim", str(path))
    with Image.open(path) as image:
        result, overlay = analyse(image, theme, surface, panel, critical)
        overlay.save(RESULTS / (name + "-regions.png"))
        # OCR only the application area, enlarged for stable recognition of small fonts.
        crop = image.crop((780, 0, 1280, 300 if panel else 140)).convert("RGB")
        crop = crop.resize((crop.width * 3, crop.height * 3))
        crop.save(RESULTS / (name + "-ocr.png"))
    expected = (["Critical notification", "The border uses the danger colour", "Normal notification", "Readable body text"]
                if critical else ["Light notification", "Theme changed without restart"])
    if panel:
        expected += ["Notifications", "Do Not Disturb", "Clear All"]
    # Save numeric evidence before invoking OCR so a timeout retains partial results.
    report = RESULTS / (name + "-metrics.json")
    report.write_text(json.dumps(result, indent=2) + "\n")
    recognised = run("tesseract", str(RESULTS / (name + "-ocr.png")), "stdout", "-l", "eng", "--psm", "6")
    result["ocr"] = {"text": recognised, "expected": expected,
                     "missing": missing_text(recognised, expected)}
    result["passed"] = result["passed"] and not result["ocr"]["missing"]
    report.write_text(json.dumps(result, indent=2) + "\n")
    for metric in result["metrics"]:
        check(metric["passed"], name + " " + metric["name"])
    check(not result["ocr"]["missing"], name + " expected text visible: " + str(result["ocr"]["missing"]))


def swaync():
    compositor = start("sway", "-c", "/opt/sandbox/sway.conf", log="sway.log")
    wait_until(lambda: list(Path(os.environ["XDG_RUNTIME_DIR"]).glob("wayland-*.lock")),
               "headless Wayland socket")
    sockets = [p for p in Path(os.environ["XDG_RUNTIME_DIR"]).glob("wayland-*") if p.is_socket()]
    check(bool(sockets) and compositor.poll() is None, "headless compositor running")
    os.environ["WAYLAND_DISPLAY"] = sockets[0].name
    config = Path.home() / ".config/swaync"
    config.mkdir(parents=True)
    shutil.copyfile("/etc/xdg/swaync/config.json", config / "config.json")
    style = Path("/etc/xdg/swaync/style.css").read_text() + '\n@import "tinct.css";\n'
    (config / "style.css").write_text(style)
    dark_surface = generate("dark")
    check(not owns_notifications(), "generation does not activate stopped daemon")
    daemon = start("swaync", log="swaync.log")
    wait_until(owns_notifications, "notification service")
    run("notify-send", "-a", "Tinct validation", "-t", "0", "Normal notification", "Readable body text")
    run("notify-send", "-a", "Critical validation", "-u", "critical", "-t", "0",
        "Critical notification", "The border uses the danger colour")
    screenshot("dark-popups", "dark", dark_surface, critical=True)
    run("swaync-client", "--open-panel", "--skip-wait")
    screenshot("dark-panel", "dark", dark_surface, panel=True, critical=True)
    light_surface = generate("light")
    check(daemon.poll() is None, "live reload preserves daemon process")
    screenshot("light-panel", "light", light_surface, panel=True, critical=True)
    run("swaync-client", "--close-panel", "--skip-wait")
    run("notify-send", "-t", "0", "Light notification", "Theme changed without restart")
    screenshot("light-popup", "light", light_surface)
    daemon.terminate()
    daemon.wait(timeout=5)
    wait_until(lambda: not owns_notifications(), "notification service stops")
    generate("light")
    check(not owns_notifications(), "reload does not activate stopped daemon")
    check((config / "style.css").read_text() == style, "user stylesheet remains unchanged")
    log = (RESULTS / "swaync.log").read_text()
    check(not any(s in log.lower() for s in ("theme parser error", "failed to load css", "css parsing error")),
          "no GTK stylesheet parsing errors")


VALIDATORS = {"swaync": swaync, "waybar": lambda: desktop.waybar(sys.modules[__name__])}


def main():
    result = {"app": sys.argv[1], "status": "failed", "checks": CHECKS}
    try:
        (RESULTS / "packages.txt").write_text(run("apk", "--no-network", "info", "-v"))
        VALIDATORS[sys.argv[1]]()
        result["status"] = "passed"
    except Exception:
        result["error"] = traceback.format_exc()
        print(result["error"], file=sys.stderr)
    finally:
        for proc in reversed(PROCESSES):
            if proc.poll() is None:
                proc.terminate()
                try:
                    proc.wait(timeout=3)
                except subprocess.TimeoutExpired:
                    proc.kill()
                    proc.wait(timeout=3)
        (RESULTS / "report.json").write_text(json.dumps(result, indent=2) + "\n")
    return 0 if result["status"] == "passed" else 1


if __name__ == "__main__":
    sys.exit(main())
