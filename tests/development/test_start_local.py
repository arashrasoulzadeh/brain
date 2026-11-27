import os
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


class LocalServerTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.bin = self.root / "bin"
        self.bin.mkdir()
        (self.root / "scripts").mkdir()
        shutil.copyfile(ROOT / "scripts/start-local.sh", self.root / "scripts/start-local.sh")
        self.env = dict(os.environ, PATH=str(self.bin))
        for name in ("HTTP_ADDRESS", "PROJECTS_FILE", "OPERATOR_USERNAME", "OPERATOR_SECRET",
                     "BRAIN_DEV_RUNTIME", "DOCKER_HOST", "DOCKER_CONTEXT", "RUNTIME_DSN", "MIGRATION_DSN"):
            self.env.pop(name, None)
        for name in ("dirname", "basename", "mkdir", "grep", "cat"):
            (self.bin / name).symlink_to(shutil.which(name))
        self.stub("uname", "echo Linux")
        self.stub("id", 'case "$1" in -u) echo 1001;; -g) echo 1001;; esac')

    def stub(self, name, body):
        path = self.bin / name
        path.write_text("#!/bin/sh\n" + body + "\n")
        path.chmod(0o755)

    def start(self, **env):
        return subprocess.run(
            [shutil.which("bash"), "scripts/start-local.sh"], cwd=self.root,
            env=dict(self.env, **env), text=True, capture_output=True, check=False,
        )

    def docker(self):
        self.stub("docker", '''case "$1" in
context) echo unix:///local/docker.sock;;
info) echo linux;;
build|run) printf '%s\\n' "$@";;
*) exit 42;;
esac''')

    def test_linux_missing_bwrap_has_actionable_diagnostic(self):
        result = self.start()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("sudo apt install bubblewrap", result.stderr)
        self.assertIn("BRAIN_DEV_RUNTIME=docker", result.stderr)

    def test_native_preserves_original_entrypoint(self):
        self.stub("bwrap", "exit 0")
        self.stub("node", "exit 0")
        self.stub("go", 'printf "%s\\n" "$@" "$HTTP_ADDRESS" "$PROJECTS_FILE"')
        self.assertIn("make setup", self.start().stderr)
        loader = self.root / "node_modules/tsx/dist/loader.mjs"
        loader.parent.mkdir(parents=True)
        loader.touch()
        result = self.start(HTTP_ADDRESS="127.0.0.1:18090")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.splitlines(), ["run", "./cmd/brain", "127.0.0.1:18090", ".local/brain.sqlite"])

    def test_macos_uses_local_linux_image_without_privileged_or_source_mount(self):
        self.stub("uname", "echo Darwin")
        self.docker()
        result = self.start(HTTP_ADDRESS="127.0.0.1:18090", PROJECTS_FILE=".local/dev.sqlite")
        self.assertEqual(result.returncode, 0, result.stderr)
        args = result.stdout.splitlines()
        self.assertIn("127.0.0.1:18090:8080", args)
        self.assertIn("PROJECTS_FILE=/workspace/.local/dev.sqlite", args)
        self.assertIn(f"type=bind,source={self.root}/.local,target=/workspace/.local", args)
        self.assertIn("1001:1001", args)
        self.assertNotIn("--privileged", args)
        self.assertNotIn("/var/run/docker.sock", result.stdout)

    def test_reject_unsafe_docker_settings(self):
        self.docker()
        for overrides in (
            {"HTTP_ADDRESS": "0.0.0.0:18082"},
            {"HTTP_ADDRESS": "127.0.0.1:99999"},
            {"PROJECTS_FILE": "outside.sqlite"},
            {"DOCKER_HOST": "tcp://remote:2375"},
            {"RUNTIME_DSN": "postgres://example"},
        ):
            with self.subTest(overrides=overrides):
                result = self.start(BRAIN_DEV_RUNTIME="docker", **overrides)
                self.assertNotEqual(result.returncode, 0)
                self.assertNotIn("build\n", result.stdout)

    def test_macos_missing_docker_and_forced_native(self):
        self.stub("uname", "echo Darwin")
        self.assertIn("Docker Desktop or Colima", self.start().stderr)
        self.assertIn("Native sandboxing requires Linux", self.start(BRAIN_DEV_RUNTIME="native").stderr)


if __name__ == "__main__":
    unittest.main()
