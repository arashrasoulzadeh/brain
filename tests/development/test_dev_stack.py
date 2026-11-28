import importlib.util
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
        with patch.object(stack, "settings", return_value=(env, "127.0.0.1", 3000)), \
                patch.object(stack.signal, "signal"), patch.object(stack.subprocess, "Popen", side_effect=launch), \
                patch.object(stack, "wait_for_brain", side_effect=wait), patch.object(stack, "stop_all"):
            stack.run()
        self.assertEqual(events, ["bash", "live", "npm"])


if __name__ == "__main__":
    unittest.main()
