#!/usr/bin/env python3
"""Opt-in integration: start the actual Brain and four OS-sandboxed Neuron processes."""

import json
import os
import signal
import subprocess
import tempfile
import time
import unittest
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


@unittest.skipUnless(os.environ.get("BRAIN_TEST_DEV_NEURONS") == "1", "requires native sandbox and Node dependencies")
class DevNeuronProcessesTest(unittest.TestCase):
    def test_four_independent_neurons_register_and_shutdown(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            binary = root / "brain"
            subprocess.run(["go", "build", "-o", str(binary), "./cmd/brain"], cwd=ROOT, check=True)
            import socket
            with socket.socket() as probe:
                probe.bind(("127.0.0.1", 0))
                port = probe.getsockname()[1]
            origin = f"http://127.0.0.1:{port}"
            env = dict(os.environ, HTTP_ADDRESS=f"127.0.0.1:{port}", PROJECTS_FILE=str(root / "test.sqlite"),
                       OPERATOR_USERNAME="ci-operator", OPERATOR_SECRET="ci-fixture-only",
                       BRAIN_LOCAL_DEV="1", BRAIN_LOCAL_MACOS="1", BRAIN_DEV_WORKERS="4", BRAIN_DEV_SLOTS="1")
            # Verify the default rather than an explicitly selected Cortex mode.
            env.pop("BRAIN_DEV_PLUGIN_EXECUTION", None)
            log_path = root / "server.log"
            with log_path.open("w+") as log:
                process = subprocess.Popen([str(binary)], cwd=ROOT, env=env, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
                try:
                    deadline = time.monotonic() + 65
                    while time.monotonic() < deadline:
                        if process.poll() is not None:
                            self.fail(log_path.read_text())
                        try:
                            request = urllib.request.Request(origin + "/_system/admin/session",
                                data=json.dumps({"username": "ci-operator", "secret": "ci-fixture-only"}).encode(),
                                headers={"Content-Type": "application/json"}, method="POST")
                            with urllib.request.urlopen(request, timeout=1) as response:
                                token = json.load(response)["session"]
                            break
                        except OSError:
                            time.sleep(.25)
                    else:
                        self.fail(log_path.read_text())
                    request = urllib.request.Request(origin + "/_system/admin/neurons", headers={"Authorization": "Bearer " + token})
                    with urllib.request.urlopen(request, timeout=3) as response:
                        status = json.load(response)
                    self.assertTrue(status["configured"])
                    neurons = status["snapshot"]["Neurons"]
                    self.assertEqual({row["ID"] for row in neurons}, {f"local-neuron-{i:02d}" for i in range(1, 5)})
                    self.assertTrue(all(row["Slots"] == 1 and {rt["Kind"] for rt in row["Runtimes"]} ==
                                        {"standalone-typescript", "plugin-typescript-v1", "user-function-v1"} for row in neurons))
                    # Reuse the same actual HTTP/storage/lifecycle suite as Linux.
                    subprocess.run(["node", str(ROOT / "scripts/test-runtime-api.mjs")], cwd=ROOT,
                                   env=dict(env, BRAIN_TEST_ORIGIN=origin), check=True, timeout=360)
                    process.send_signal(signal.SIGTERM)
                    self.assertEqual(process.wait(timeout=15), 0, log_path.read_text())
                finally:
                    try:
                        os.killpg(process.pid, signal.SIGKILL)
                    except ProcessLookupError:
                        pass
                    process.wait(timeout=5)


if __name__ == "__main__":
    unittest.main()
