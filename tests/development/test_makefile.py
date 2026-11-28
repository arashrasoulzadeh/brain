import os
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


class DevelopmentMakefileTest(unittest.TestCase):
    def run_make(self, *args, env=None):
        return subprocess.run(
            ["make", "--no-print-directory", *args], cwd=ROOT, env=env,
            capture_output=True, text=True, check=False,
        )

    def test_server_and_admin_dispatch_to_existing_entrypoints(self):
        with tempfile.TemporaryDirectory() as directory:
            bin_dir = Path(directory)
            for name in ("bash", "npm"):
                command = bin_dir / name
                command.write_text(
                    '#!/bin/sh\nprintf "%s\\n" "$0" "$@" "${BRAIN_API_ORIGIN:-}"\n',
                    encoding="utf-8",
                )
                command.chmod(0o755)
            env = dict(os.environ, PATH=f"{bin_dir}{os.pathsep}{os.environ['PATH']}")
            server = self.run_make("dev", "server", env=env)
            self.assertEqual(server.returncode, 0, server.stderr)
            self.assertIn(f"{bin_dir}/bash\nscripts/start-local.sh\n", server.stdout)
            admin = self.run_make("dev", "admin", "ADMIN_PORT=3100", "ADMIN_HOST=localhost",
                                  "BRAIN_API_ORIGIN=http://127.0.0.1:18090", env=env)
            self.assertEqual(admin.returncode, 0, admin.stderr)
            self.assertIn(
                f"{bin_dir}/npm\n--prefix\nadmin-nextjs\nrun\ndev\n--\n--hostname\nlocalhost\n"
                "--port\n3100\nhttp://127.0.0.1:18090\n", admin.stdout,
            )

    def test_defaults_setup_and_invalid_dev_commands(self):
        admin = self.run_make("-n", "dev", "admin")
        self.assertEqual(admin.returncode, 0, admin.stderr)
        self.assertIn('BRAIN_API_ORIGIN="http://127.0.0.1:18082"', admin.stdout)
        self.assertIn('--hostname "127.0.0.1" --port "3000"', admin.stdout)
        setup = self.run_make("-n", "setup")
        self.assertEqual(setup.returncode, 0, setup.stderr)
        self.assertIn("npm ci\n", setup.stdout)
        self.assertIn("npm --prefix admin-nextjs ci", setup.stdout)
        stack = self.run_make("-n", "dev", "ADMIN_PORT=3100")
        self.assertEqual(stack.returncode, 0, stack.stderr)
        self.assertIn('ADMIN_PORT="3100"', stack.stdout)
        self.assertIn("python3 scripts/dev-stack.py", stack.stdout)
        self.assertIn('BRAIN_DEV_WORKERS="4"', stack.stdout)
        self.assertIn('BRAIN_DEV_SLOTS="1"', stack.stdout)
        custom = self.run_make("-n", "dev", "WORKERS=2", "WORKER_SLOTS=3")
        self.assertIn('BRAIN_DEV_WORKERS="2"', custom.stdout)
        self.assertIn('BRAIN_DEV_SLOTS="3"', custom.stdout)
        server = self.run_make("-n", "dev", "server", "WORKERS=0", "WORKER_SLOTS=2")
        self.assertEqual(server.returncode, 0, server.stderr)
        self.assertIn('BRAIN_DEV_WORKERS="0"', server.stdout)
        self.assertIn('BRAIN_DEV_SLOTS="2"', server.stdout)
        for args in [("dev", "server", "admin")]:
            result = self.run_make(*args)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("Usage: make dev, make dev server, or make dev admin", result.stderr)


if __name__ == "__main__":
    unittest.main()
