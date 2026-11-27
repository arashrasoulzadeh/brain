"""Run the actual Brain executable; GOCOVERDIR is inherited in coverage CI."""
import os
import signal
import socket
import time
from urllib.error import HTTPError
from urllib.request import urlopen
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]


class CLITest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.temp = tempfile.TemporaryDirectory()
        cls.addClassCleanup(cls.temp.cleanup)
        cls.binary = Path(os.environ.get("BRAIN_TEST_CLI", Path(cls.temp.name) / "brain")).resolve()
        if "BRAIN_TEST_CLI" not in os.environ:
            subprocess.run(["go", "build", "-o", str(cls.binary), "./cmd/brain"], cwd=ROOT, check=True)

    def cli(self, *args, expected=0, contains="", **env):
        environment = os.environ.copy()
        for key in ("HTTP_ADDRESS", "OPERATOR_USERNAME", "OPERATOR_SECRET", "PROJECTS_FILE", "RUNTIME_DSN", "MIGRATION_DSN"):
            environment.pop(key, None)
        environment.update(env)
        result = subprocess.run([str(self.binary), *map(str, args)], cwd=ROOT, env=environment,
                                capture_output=True, text=True, timeout=30)
        self.assertEqual(result.returncode, expected, result.stdout + result.stderr)
        self.assertIn(contains, result.stdout + result.stderr)
        return result.stdout

    def test_health_only_serves_and_shuts_down(self):
        with socket.socket() as listener:
            listener.bind(("127.0.0.1", 0))
            port = listener.getsockname()[1]
        env = os.environ.copy()
        env.update(HTTP_ADDRESS=f"127.0.0.1:{port}", RUNTIME_DSN="", MIGRATION_DSN="")
        process = subprocess.Popen([str(self.binary), "health-only"], cwd=ROOT, env=env,
                                   stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        try:
            for _ in range(100):
                if process.poll() is not None:
                    self.fail(f"health server exited: {process.communicate()}")
                try:
                    with urlopen(f"http://127.0.0.1:{port}/livez", timeout=1) as response:
                        self.assertEqual(response.status, 200)
                    break
                except OSError:
                    time.sleep(0.05)
            else:
                self.fail("health server never became live")
            with self.assertRaises(HTTPError) as error:
                urlopen(f"http://127.0.0.1:{port}/readyz", timeout=1)
            self.assertEqual(error.exception.code, 503)
            process.send_signal(signal.SIGTERM)
            stdout, stderr = process.communicate(timeout=10)
            self.assertEqual(process.returncode, 0, stdout + stderr)
        finally:
            if process.poll() is None:
                process.kill()
                process.communicate()

    def test_health_only_reports_listen_failure(self):
        self.cli("health-only", expected=1, contains="missing port in address",
                 HTTP_ADDRESS="invalid-address")

    def test_every_plugin_contract_in_both_formats(self):
        for manifest in sorted(ROOT.glob("plugins/*/plugin.yaml")) + sorted(ROOT.glob("examples/*/plugin.yaml")):
            with self.subTest(manifest=manifest):
                self.assertTrue(self.cli("plugin-contract", manifest, "markdown").strip())
                self.assertTrue(self.cli("plugin-contract", manifest, "typescript").strip())

    def test_invalid_contract_requests(self):
        self.cli("plugin-contract", expected=1, contains="usage:")
        self.cli("plugin-contract", ROOT / "absent.yaml", "markdown", expected=1)
        self.cli("plugin-contract", ROOT / "plugins/arena/plugin.yaml", "unknown", expected=1, contains="unknown contract format")

    def test_bootstrap_commands(self):
        self.cli("bootstrap-dashboard", contains="dashboard plugin")
        self.cli("bootstrap-operator", OPERATOR_USERNAME="fixture", OPERATOR_SECRET="fixture-secret", contains="bootstrapped")
        self.cli("bootstrap-operator", expected=1)

    def test_object_manifest(self):
        path = Path(self.temp.name) / "objects.json"
        path.write_text('{"objects":[{"project_id":"p","object_key":"file","version_id":"v1"}]}', encoding="utf-8")
        self.cli("verify-object-manifest", path, contains="object manifest verified")
        path.write_text("invalid", encoding="utf-8")
        self.cli("verify-object-manifest", path, expected=1, contains="invalid object version manifest")

    def test_production_configuration_rejects_unsafe_startup(self):
        self.cli(expected=1, contains="HTTP_ADDRESS is required")
        self.cli(expected=1, contains="OPERATOR_USERNAME and OPERATOR_SECRET are required", HTTP_ADDRESS="127.0.0.1:0")
        env = dict(HTTP_ADDRESS="127.0.0.1:0", OPERATOR_USERNAME="fixture", OPERATOR_SECRET="fixture-secret")
        self.cli(expected=1, contains="PostgreSQL application storage is not wired", RUNTIME_DSN="fixture", **env)
        for value in ("", ":memory:", "file:fixture"):
            self.cli(expected=1, contains="PROJECTS_FILE must name an explicit durable SQLite file", PROJECTS_FILE=value, **env)


if __name__ == "__main__":
    unittest.main()
