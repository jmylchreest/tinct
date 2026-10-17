#!/usr/bin/env python3
"""Run real application validators with Docker or rootless Podman."""
import argparse
import json
import pathlib
import shutil
import subprocess
import sys
import uuid

HERE = pathlib.Path(__file__).resolve().parent


def command(argv, timeout=60, **kwargs):
    return subprocess.run(argv, check=True, timeout=timeout, **kwargs)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--engine", choices=("docker", "podman"), default="docker")
    parser.add_argument("--binary", required=True, type=pathlib.Path)
    parser.add_argument("--app", choices=("swaync", "waybar", "fuzzel", "foot"), default="swaync")
    parser.add_argument("--artifacts", required=True, type=pathlib.Path)
    args = parser.parse_args()
    if not args.binary.is_file():
        parser.error("--binary must be a statically linked Linux tinct executable")
    if not shutil.which(args.engine):
        parser.error(f"{args.engine} is not installed")
    args.artifacts.mkdir(parents=True, exist_ok=True)
    if any(args.artifacts.iterdir()):
        parser.error("--artifacts must be empty to avoid mixing results from different runs")
    name = "tinct-validator-" + uuid.uuid4().hex[:12]
    image = name + ":test"
    created = False
    result = {"app": args.app, "status": "failed"}
    try:
        with (args.artifacts / "build.log").open("w") as log:
            command([args.engine, "build", "-t", image, str(HERE)], timeout=600,
                    stdout=log, stderr=subprocess.STDOUT)
        created = True  # Even a timed-out create may have allocated the named container.
        command([args.engine, "create", "--name", name, "--network", "none",
                 "--cap-drop", "ALL", "--security-opt", "no-new-privileges",
                 "--pids-limit", "256", "--memory", "1g", image, args.app])
        command([args.engine, "cp", str(args.binary.resolve()), name + ":/usr/local/bin/tinct"])
        with (args.artifacts / "container.log").open("w") as log:
            command([args.engine, "start", "--attach", name], timeout=120,
                    stdout=log, stderr=subprocess.STDOUT)
        # Check explicitly: runtime start/attach exit semantics vary by engine.
        state = command([args.engine, "inspect", "--format", "{{.State.ExitCode}}", name],
                        capture_output=True, text=True).stdout.strip()
        if state != "0":
            raise RuntimeError(f"validator exited with {state}")
        result["status"] = "passed"
    except (subprocess.SubprocessError, OSError, RuntimeError) as error:
        result["error"] = str(error)
        print(error, file=sys.stderr)
    finally:
        if created:
            # Stop first on timeout; retain partial logs and screenshots before removal.
            for argv in ([args.engine, "stop", "--time", "2", name],
                         [args.engine, "cp", name + ":/tmp/results/.", str(args.artifacts)],
                         [args.engine, "rm", "--force", name]):
                try:
                    command(argv)
                except (subprocess.SubprocessError, OSError) as error:
                    result["status"] = "failed"
                    result.setdefault("cleanup_errors", []).append(str(error))
        try:
            command([args.engine, "image", "rm", image])
        except (subprocess.SubprocessError, OSError):
            pass  # A failed build may never create the image.
        if result["status"] == "passed":
            try:
                report = json.loads((args.artifacts / "report.json").read_text())
                if report.get("status") != "passed" or report.get("app") != args.app or not report.get("checks"):
                    raise ValueError("missing successful application checks")
            except (OSError, ValueError) as error:
                result.update(status="failed", error=str(error))
        (args.artifacts / "runner.json").write_text(json.dumps(result, indent=2) + "\n")
    return 0 if result["status"] == "passed" else 1


if __name__ == "__main__":
    sys.exit(main())
