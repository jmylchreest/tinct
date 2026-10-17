import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
import run


class RunnerTests(unittest.TestCase):
    def exercise(self, failure=None, missing_report=False):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            binary = root / "tinct"
            binary.touch()
            artifacts = root / "results"
            calls = []

            def fake(argv, **kwargs):
                calls.append(argv)
                if argv[1] == failure:
                    raise subprocess.TimeoutExpired(argv, 1)
                if argv[1] == "cp" and ":/tmp/results/." in argv[2] and not missing_report:
                    (artifacts / "report.json").write_text(json.dumps({
                        "app": "swaync", "status": "passed", "checks": ["real test"]}))
                return subprocess.CompletedProcess(argv, 0, stdout="0\n")

            with patch("sys.argv", ["run.py", "--binary", str(binary), "--artifacts", str(artifacts)]), \
                 patch.object(run.shutil, "which", return_value="/usr/bin/docker"), \
                 patch.object(run, "command", side_effect=fake):
                status = run.main()
            result = json.loads((artifacts / "runner.json").read_text())
            return status, result, calls

    def test_timeout_retains_artifacts_then_removes_container(self):
        status, result, calls = self.exercise("start")
        self.assertEqual(status, 1)
        self.assertEqual(result["status"], "failed")
        operations = [argv[1] for argv in calls]
        self.assertEqual(operations[-4:], ["stop", "cp", "rm", "image"])
        self.assertIn(":/tmp/results/.", calls[-3][2])

    def test_create_timeout_still_cleans_up(self):
        status, result, calls = self.exercise("create")
        self.assertEqual(status, 1)
        self.assertEqual([argv[1] for argv in calls][-4:], ["stop", "cp", "rm", "image"])

    def test_build_failure_does_not_touch_nonexistent_container(self):
        status, result, calls = self.exercise("build")
        self.assertEqual(status, 1)
        self.assertEqual([argv[1] for argv in calls], ["build", "image"])

    def test_success_is_isolated(self):
        status, result, calls = self.exercise()
        self.assertEqual(status, 0)
        create = next(argv for argv in calls if argv[1] == "create")
        self.assertEqual(create[create.index("--network") + 1], "none")
        self.assertEqual(create[create.index("--cap-drop") + 1], "ALL")
        self.assertNotIn("--privileged", create)
        self.assertNotIn("--volume", create)

    def test_cleanup_failure_is_not_success(self):
        status, result, calls = self.exercise("stop")
        self.assertEqual(status, 1)
        self.assertTrue(result["cleanup_errors"])
        self.assertIn("rm", [argv[1] for argv in calls])

    def test_missing_report_is_not_success(self):
        status, result, calls = self.exercise(missing_report=True)
        self.assertEqual(status, 1)
        self.assertEqual(result["status"], "failed")

    def test_coverage_inventory(self):
        here = Path(__file__).resolve().parent
        outputs = here.parents[1] / "internal/plugin/output"
        expected = {p.name for p in outputs.iterdir() if p.is_dir()} - {"shared", "template"}
        coverage = json.loads((here / "coverage.json").read_text())
        self.assertEqual(set(coverage), expected)
        for app, entry in coverage.items():
            self.assertIn(entry["status"], ("planned", "implemented"))
            if entry["status"] == "implemented":
                self.assertIn(app, ("swaync", "waybar"))


if __name__ == "__main__":
    unittest.main()
