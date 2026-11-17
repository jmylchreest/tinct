---
sidebar_position: 6
---

# Application validation

The repository's application sandbox tests run the real Tinct CLI and real applications in an isolated headless Wayland session. Current adapters cover SwayNC, Waybar, Fuzzel and Foot. Other plugins do not yet have visual coverage.

From a Linux checkout with Go, Python 3.10+ and Docker or rootless Podman:

```sh
CGO_ENABLED=0 go build -o /tmp/tinct-validator ./cmd/tinct
python3 tests/sandbox/run.py --engine podman --binary /tmp/tinct-validator \
  --app foot --artifacts /tmp/tinct-foot-results
```

Use an empty artifact directory for each run. Replace `podman` with `docker` to use Docker. Image building requires network access; application execution has no network or access to your desktop session.

The GitHub sandbox workflow runs the same adapters on standard Ubuntu runners without a GPU. Each run retains application logs, screenshots, OCR results and numeric colour/contrast measurements. A successful command alone is not sufficient: the application must display the expected theme and text. Review the screenshots for layout and clipping too; these checks do not prove every possible user layout is correct.

Foot's validator checks generation without live updates, an explicit OSC update on the same running process, and a fresh process loading the saved palette. SwayNC and Waybar exercise live reloads; Fuzzel exercises selection, filtering and accepted output in dark and light themes.

See the [sandbox guide](https://github.com/jmylchreest/tinct/blob/main/tests/sandbox/README.md) for isolation details, supported checks and how to add an adapter. The [coverage inventory](https://github.com/jmylchreest/tinct/blob/main/tests/sandbox/coverage.json) distinguishes implemented validators from planned coverage.
