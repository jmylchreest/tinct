# Application sandbox validators

Run real applications against files produced by the actual Tinct CLI. The first
adapter is SwayNC; this is an extensible suite, not a claim that all output plugins
have visual coverage. See `coverage.json` for the explicit inventory.

## Run locally

Requires Linux, Python 3.10+, Go, and Docker or rootless Podman:

```sh
CGO_ENABLED=0 go build -o /tmp/tinct-validator ./cmd/tinct
python3 tests/sandbox/run.py --engine podman --binary /tmp/tinct-validator \
  --app swaync --artifacts /tmp/tinct-sandbox-results
```

Use a new/empty artifact directory for each run. The binary must match the Linux
container architecture and be executable. The runner builds an Alpine 3.24 image;
package versions are retained in `packages.txt`. Dependencies can receive Alpine
patch updates: screenshots are checked for semantic colours, not exact golden
image equality.

The runtime has no network, host desktop sockets, home-directory mounts, GPU,
secrets, Docker socket, or elevated capabilities. A non-root user starts a private
D-Bus session and headless Sway compositor with software rendering. Package
installation during image build requires network access. Commands use argument
arrays, deadlines, and cleanup on failure. This is application isolation for trusted
project tests, not a security boundary for deliberately hostile code.

## What SwayNC checks

- Generate dark and light styles from sparse four-colour input with the real CLI.
- Start the actual GTK4 application, send normal and critical notifications, and
  capture popup and control-panel screenshots.
- Assert expected background and critical-border pixels appear in the rendered
  image; a blank desktop or failed theme switch fails the check.
- Reload into the same live daemon process and verify the light rendering.
- Generate while stopped without activating the notification service.
- Preserve the existing user stylesheet and reject GTK CSS parsing errors.

`runner.json` reports orchestration status; `report.json` records application
checks. Build/command/application logs and partial screenshots survive failures.
Review screenshots for layout, clipping and readability: colour assertions do not
replace human visual review. A failed run exits nonzero.

## GitHub runners

`.github/workflows/sandbox.yml` runs this on standard `ubuntu-24.04` hosted runners
using Docker. No GPU, nested VM, privileged container or desktop login is required.
The workflow uses read-only repository permissions, no secrets, a job deadline,
and always uploads artifacts. Do not use the restricted `ubuntu-slim` runner.
The same harness runs with rootless Podman locally.

## Add an adapter

Add a validator to `scenario.py` and its explicit CLI choice in `run.py`, install
its runtime in the Dockerfile, update `coverage.json`, and add it to the workflow
matrix. Use a fresh container per app. Validate through the real application's
parser or renderer, including generation, reload and failure behaviour where
available. Keep native parser checks distinct from visual checks. Do not mark
missing applications or unsupported environments as passing.

Wayland apps (Waybar, Fuzzel, Foot), X11 apps (Dunst through Xvfb), terminal apps
(tmux, Neovim), GTK/Qt fixtures, and compositor-specific sessions need separate
adapters. GNOME Shell/KDE sessions and applications unavailable in Alpine need
separate images or runner arrangements. Wallpaper adapters need real compositor
protocol tests; mock binary success alone does not prove they work.

## Quantitative screenshot evidence

Each screenshot now produces `*-metrics.json` with measured values, thresholds,
pass/fail results and region coordinates, plus `*-regions.png` showing checked
bounds. Checks measure notification surface and foreground pixel fractions,
critical-border coverage, panel background/header coverage and preservation of
the desktop outside the application. Foreground/background contrast must reach
4.5:1; this checks the expected solid colours which must also occur in the image,
not every antialiased text pixel or overall accessibility compliance.

Tesseract reads enlarged application crops and requires the expected notification
titles, body text and panel controls. OCR text and missing phrases are retained
in the metric report. OCR is an additional visibility check, not proof that every
glyph is unclipped or every possible overlap is absent.

These regions target the fixed 1280x900 default SwayNC layout in this fixture.
Application layout changes may require reviewing and updating them. There is no
whole-screen similarity score or golden screenshot: unchanged desktop pixels
cannot compensate for a missing panel or notification. Synthetic negative tests
cover blank screens, hidden text, absent panels/borders and displaced content.
To run the Python tests locally, install Pillow 12.3.0 in a virtual environment,
then run `python -m unittest discover -s tests/sandbox -p 'test_*.py' -v`.

## Waybar

`--app waybar` starts a real bar with fixed title, media and clock labels. It
imports Tinct's generated example stylesheet, verifies bar/module colours, text
contrast and OCR, then regenerates a light theme and checks the same Waybar
process renders it through Tinct's SIGUSR2 hook. The user stylesheet and desktop
region must remain unchanged. The fixture fixes module content and font size;
it does not test hardware-backed modules or every possible user layout.

## Fuzzel

`--app fuzzel` validates the generated include through Fuzzel's native parser,
then renders a fixed three-entry menu with the generated opacity preserved.
Dark/light screenshots measure selected/unselected text and background colours,
contrast and OCR. Virtual Wayland keyboard events move the selection and filter
the list; accepting must return the expected entry. Fuzzel has no live theme
reload, so each theme starts a new process. No host keyboard is involved.
