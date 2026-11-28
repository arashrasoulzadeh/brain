#!/usr/bin/env python3
"""Run the native development services as one supervised process group per service."""

import os
import signal
import subprocess
import sys
import time
from pathlib import Path
from urllib.parse import urlsplit

ROOT = Path(__file__).resolve().parents[1]


def settings():
    env = os.environ.copy()
    origin = urlsplit(env.get("BRAIN_API_ORIGIN", "http://127.0.0.1:18082"))
    host = env.get("ADMIN_HOST", "127.0.0.1")
    port = int(env.get("ADMIN_PORT", "3000"))
    if (origin.scheme != "http" or origin.hostname != "127.0.0.1"
            or not origin.port or origin.path not in ("", "/")
            or origin.query or origin.fragment or origin.username or origin.password
            or host not in ("127.0.0.1", "localhost", "::1")
            or not 1 <= port <= 65535):
        raise ValueError("make dev requires loopback HTTP addresses and valid ports")
    address = f"127.0.0.1:{origin.port}"
    if env.get("HTTP_ADDRESS", address) != address:
        raise ValueError("HTTP_ADDRESS must match BRAIN_API_ORIGIN")
    workers = int(env.get("BRAIN_DEV_WORKERS", "4"))
    slots = int(env.get("BRAIN_DEV_SLOTS", "1"))
    if not 0 <= workers <= 32 or not 1 <= slots <= 8:
        raise ValueError("WORKERS must be 0–32 and WORKER_SLOTS must be 1–8")
    env["BRAIN_LOCAL_DEV"] = "1"
    env["BRAIN_DEV_WORKERS"] = str(workers)
    env["BRAIN_DEV_SLOTS"] = str(slots)
    env["HTTP_ADDRESS"] = address
    env["BRAIN_API_ORIGIN"] = f"http://{address}"
    return env, host, port


def stop_all(children):
    # Kill groups, including go run's child binary and npm's Next.js child.
    for sig in (signal.SIGTERM, signal.SIGKILL):
        for _, child in children:
            try:
                os.killpg(child.pid, sig)
            except ProcessLookupError:
                pass
        if sig == signal.SIGTERM:
            deadline = time.monotonic() + 12
            while time.monotonic() < deadline:
                if all(child.poll() is not None for _, child in children):
                    break
                time.sleep(0.05)
    for _, child in children:
        child.wait()


def run():
    env, host, port = settings()
    commands = (
        ("server", ["bash", "scripts/start-local.sh"]),
        ("admin", ["npm", "--prefix", "admin-nextjs", "run", "dev", "--",
                   "--hostname", host, "--port", str(port)]),
    )
    children = []
    stopped = False

    def stop(_signal, _frame):
        nonlocal stopped
        stopped = True

    signal.signal(signal.SIGINT, stop)
    signal.signal(signal.SIGTERM, stop)
    try:
        for name, command in commands:
            if stopped:
                return 0
            print(f"Starting {name}", flush=True)
            children.append((name, subprocess.Popen(command, cwd=ROOT, env=env,
                                                   start_new_session=True)))
        print(f"Admin: http://{host}:{port} | Brain: {env['BRAIN_API_ORIGIN']}", flush=True)
        print(f"Managed Neurons: {env['BRAIN_DEV_WORKERS']} requested · {env['BRAIN_DEV_SLOTS']} slots each", flush=True)
        print("Ctrl+C stops the entire development stack.", flush=True)
        while not stopped:
            for name, child in children:
                status = child.poll()
                if status is not None:
                    print(f"{name} exited ({status}); stopping development stack.", file=sys.stderr, flush=True)
                    return status if status > 0 else 1
            time.sleep(0.1)
        return 0
    finally:
        stop_all(children)


if __name__ == "__main__":
    try:
        sys.exit(run())
    except (OSError, ValueError) as error:
        print(f"Development startup failed: {error}", file=sys.stderr)
        sys.exit(1)
