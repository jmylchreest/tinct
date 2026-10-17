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
