import importlib.util
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch
from urllib.error import URLError


ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("dev_stack", ROOT / "scripts/dev-stack.py")
stack = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(stack)


class DevStackStartupTest(unittest.TestCase):
    def test_waits_until_brain_liveness_succeeds(self):
        child = MagicMock()
        child.poll.return_value = None
        response = MagicMock()
        response.__enter__.return_value = response
        response.status = 200
        response.read.return_value = b"ok\n"
        with patch.object(stack, "build_opener") as build, patch.object(stack.time, "sleep"):
            build.return_value.open.side_effect = [URLError("connection refused"), response]
            self.assertTrue(stack.wait_for_brain(child, "http://127.0.0.1:18082", lambda: False))
            self.assertEqual(build.return_value.open.call_count, 2)
            build.return_value.open.assert_called_with("http://127.0.0.1:18082/_system/livez", timeout=1)

    def test_fails_when_backend_exits(self):
        child = MagicMock()
        child.poll.return_value = 1
        with self.assertRaisesRegex(OSError, "Brain exited"):
            stack.wait_for_brain(child, "http://127.0.0.1:18082", lambda: False)

    def test_timeout_and_interrupt(self):
        child = MagicMock()
        child.poll.return_value = None
        with patch.object(stack, "build_opener") as build:
            build.return_value.open.side_effect = URLError("connection refused")
            with self.assertRaisesRegex(OSError, "startup timeout"):
                stack.wait_for_brain(child, "http://127.0.0.1:18082", lambda: False, timeout=0)
            self.assertFalse(stack.wait_for_brain(child, "http://127.0.0.1:18082", lambda: True))

    def test_admin_starts_only_after_brain_is_live(self):
        server, admin = MagicMock(), MagicMock()
        server.poll.return_value = None
        admin.poll.return_value = 0
        env = {"BRAIN_API_ORIGIN": "http://127.0.0.1:18082", "BRAIN_DEV_WORKERS": "0", "BRAIN_DEV_SLOTS": "1"}
        events = []
        def launch(command, **_kwargs):
            events.append(command[0])
            return server if command[0] == "bash" else admin
        def wait(child, origin, stopped):
            self.assertIs(child, server)
            events.append("live")
            return True
        with patch.object(stack.subprocess, "Popen", side_effect=launch), \
                patch.object(stack, "wait_for_brain", side_effect=wait), patch.object(stack, "stop_all") as cleanup:
            self.assertEqual(stack.run_generation(env, "127.0.0.1", 3000, lambda: False, lambda: False), "failed")
            cleanup.assert_called_once_with([("server", server), ("admin", admin)])
        self.assertEqual(events, ["bash", "live", "npm"])


    def test_watcher_detects_edits_additions_and_deletions_but_not_generated_files(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name in ("internal", "admin-nextjs/app", "admin-nextjs/.next-dev", "plugins/node_modules", ".local"):
                (root / name).mkdir(parents=True)
            source = root / "internal/main.go"
            source.write_text("first")
            now = [0.0]
            watcher = stack.SourceWatcher(root, interval=0.5, debounce=0.75, clock=lambda: now[0])

            def scan(at):
                now[0] = at
                return watcher.changed()

            (root / ".local/brain.sqlite").write_text("data")
            (root / "admin-nextjs/.next-dev/build.js").write_text("generated")
            (root / "plugins/node_modules/cache.js").write_text("dependency")
            self.assertFalse(scan(1))
            self.assertFalse(scan(2))
            source.write_text("changed source")
            self.assertFalse(scan(3))
            new_file = root / "admin-nextjs/app/new.tsx"
            new_file.write_text("new component")
            self.assertFalse(scan(3.5))
            self.assertFalse(scan(4))
            self.assertTrue(scan(4.5))
            watcher.reset()
            source.unlink()
            self.assertFalse(scan(5))
            self.assertTrue(scan(6))

    def test_source_change_stops_entire_generation(self):
        server, admin = MagicMock(), MagicMock()
        server.poll.return_value = None
        admin.poll.return_value = None
        env = {"BRAIN_API_ORIGIN": "http://127.0.0.1:18082", "BRAIN_DEV_WORKERS": "4", "BRAIN_DEV_SLOTS": "1"}
        with patch.object(stack.subprocess, "Popen", side_effect=[server, admin]), \
                patch.object(stack, "wait_for_brain", return_value=True), patch.object(stack, "stop_all") as cleanup:
            outcome = stack.run_generation(env, "127.0.0.1", 3000, lambda: False, MagicMock(side_effect=[False, False, True]))
            self.assertEqual(outcome, "changed")
            cleanup.assert_called_once_with([("server", server), ("admin", admin)])

    def test_edit_during_startup_does_not_launch_admin(self):
        server = MagicMock()
        env = {"BRAIN_API_ORIGIN": "http://127.0.0.1:18082"}
        with patch.object(stack.subprocess, "Popen", return_value=server) as launch, \
                patch.object(stack, "wait_for_brain", side_effect=lambda child, origin, interrupted: not interrupted()), \
                patch.object(stack, "stop_all") as cleanup:
            outcome = stack.run_generation(env, "127.0.0.1", 3000, lambda: False, MagicMock(side_effect=[False, True]))
            self.assertEqual(outcome, "changed")
            launch.assert_called_once()
            cleanup.assert_called_once_with([("server", server)])

    def test_supervisor_restarts_after_edits_and_retries_failed_build(self):
        env = {"BRAIN_API_ORIGIN": "http://127.0.0.1:18082"}
        callbacks = {}

        def signal_handler(sig, callback):
            callbacks[sig] = callback

        def generation(*args):
            if runner.call_count == 1:
                return "changed"
            if runner.call_count == 2:
                return "failed"
            callbacks[stack.signal.SIGINT](None, None)
            return "stopped"

        with patch.object(stack, "settings", return_value=(env, "127.0.0.1", 3000)), \
                patch.object(stack.signal, "signal", side_effect=signal_handler), \
                patch.object(stack, "SourceWatcher") as watcher, \
                patch.object(stack, "run_generation", side_effect=generation) as runner:
            watcher.return_value.changed.return_value = True
            self.assertEqual(stack.run(), 0)
            self.assertEqual(runner.call_count, 3)
            self.assertEqual(watcher.return_value.reset.call_count, 2)
            watcher.return_value.changed.assert_called_once()


if __name__ == "__main__":
    unittest.main()
