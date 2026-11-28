#!/usr/bin/env python3
"""Run the native development services as one supervised process group per service."""

import os
import signal
import subprocess
import sys
import time
from pathlib import Path
from urllib.parse import urlsplit
from urllib.request import ProxyHandler, build_opener
from urllib.error import URLError

ROOT = Path(__file__).resolve().parents[1]


# Watch runtime and admin sources, including untracked/new files. Do not scan
# dependency trees, generated output, or local application state.
WATCH_DIRS = ("cmd", "internal", "plugins", "proto", "protocol", "scripts", "admin-nextjs")
WATCH_FILES = ("go.mod", "go.sum", "package.json", "package-lock.json", "tsconfig.json", "vendor/modules.txt")
IGNORED_DIRS = {".git", ".local", ".next", ".next-dev", ".turbo", ".cache", ".generated-docs",
                ".generated-site", "node_modules", "__pycache__", "coverage", "dist", "build", "out"}
IGNORED_SUFFIXES = (".pyc", ".pyo", ".tsbuildinfo", ".log", ".sqlite", ".sqlite3", ".db",
                    ".sqlite-wal", ".sqlite-shm", ".sqlite3-wal", ".sqlite3-shm", ".db-wal", ".db-shm", "~")


def source_snapshot(root):
    snapshot = {}

    def record(path):
        if path.name == ".DS_Store" or path.name.endswith(IGNORED_SUFFIXES):
            return
        try:
            stat = path.lstat()
        except FileNotFoundError:
            return  # Atomic editor replacement or concurrent deletion.
        snapshot[path.relative_to(root).as_posix()] = (
            stat.st_mtime_ns, stat.st_ctime_ns, stat.st_size, stat.st_ino)

    for name in WATCH_FILES:
        record(root / name)
    for name in WATCH_DIRS:
        for directory, dirs, names in os.walk(root / name, followlinks=False):
            dirs[:] = [entry for entry in dirs if entry not in IGNORED_DIRS]
            for filename in names:
                record(Path(directory) / filename)
    return snapshot


class SourceWatcher:
    """Poll without an extra dependency; coalesce bursts of editor writes."""

    def __init__(self, root, interval=0.5, debounce=0.75, clock=time.monotonic):
        self.root = root
        self.interval = interval
        self.debounce = debounce
        self.clock = clock
        self.reset()

    def reset(self):
        self.snapshot = source_snapshot(self.root)
        self.last_scan = self.clock()
        self.pending_since = None

    def changed(self):
        now = self.clock()
        if now - self.last_scan < self.interval:
            return False
        current = source_snapshot(self.root)
        self.last_scan = now
        if current != self.snapshot:
            self.snapshot = current
            self.pending_since = now
        return self.pending_since is not None and now - self.pending_since >= self.debounce


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


def wait_for_brain(child, origin, stopped, timeout=300):
    """Wait for HTTP liveness, not application readiness (WORKERS=0 is valid)."""
    opener = build_opener(ProxyHandler({}))
    deadline = time.monotonic() + timeout
    next_notice = time.monotonic() + 10
    while not stopped():
        status = child.poll()
        if status is not None:
            raise OSError(f"Brain exited ({status}) before HTTP startup; see server output above")
        try:
            with opener.open(origin + "/_system/livez", timeout=1) as response:
                if response.status == 200 and response.read(16) == b"ok\n":
                    if child.poll() is not None:
                        raise OSError("Brain exited during HTTP startup")
                    return True
        except (URLError, TimeoutError, ConnectionError):
            pass
        now = time.monotonic()
        if now >= deadline:
            raise OSError("Brain did not become live before the development startup timeout; see server output above")
        if now >= next_notice:
            print("Waiting for Brain to finish building and starting Neurons…", flush=True)
            next_notice = now + 10
        time.sleep(0.25)
    return False


def run_generation(env, host, port, stopped, changed):
    commands = (
        ("server", ["bash", "scripts/start-local.sh"]),
        ("admin", ["npm", "--prefix", "admin-nextjs", "run", "dev", "--",
                   "--hostname", host, "--port", str(port)]),
    )
    children = []
    restart = False

    def interrupted():
        nonlocal restart
        if stopped():
            return True
        restart = restart or changed()
        return restart

    try:
        for name, command in commands:
            if interrupted():
                return "stopped" if stopped() else "changed"
            print(f"Starting {name}", flush=True)
            child = subprocess.Popen(command, cwd=ROOT, env=env, start_new_session=True)
            children.append((name, child))
            if name == "server":
                print(f"Waiting for Brain at {env['BRAIN_API_ORIGIN']} before starting admin.", flush=True)
                if not wait_for_brain(child, env["BRAIN_API_ORIGIN"], interrupted):
                    return "stopped" if stopped() else "changed"
        print(f"Admin: http://{host}:{port} | Brain: {env['BRAIN_API_ORIGIN']}", flush=True)
        print(f"Managed Neurons: {env['BRAIN_DEV_WORKERS']} requested · {env['BRAIN_DEV_SLOTS']} slots each", flush=True)
        while not interrupted():
            for name, child in children:
                status = child.poll()
                if status is not None:
                    print(f"{name} exited ({status}); stopping development services.", file=sys.stderr, flush=True)
                    return "failed"
            time.sleep(0.1)
        return "stopped" if stopped() else "changed"
    except OSError as error:
        print(f"Development generation failed: {error}", file=sys.stderr, flush=True)
        return "failed"
    finally:
        stop_all(children)


def run():
    env, host, port = settings()
    stopped = False

    def stop(_signal, _frame):
        nonlocal stopped
        stopped = True

    signal.signal(signal.SIGINT, stop)
    signal.signal(signal.SIGTERM, stop)
    watcher = SourceWatcher(ROOT)
    print("Watching runtime and admin source files. Ctrl+C stops the entire development stack.", flush=True)
    while not stopped:
        outcome = run_generation(env, host, port, lambda: stopped, watcher.changed)
        if stopped or outcome == "stopped":
            break
        if outcome == "failed":
            # Never spin on a compiler error or broken configuration. A save
            # triggers another complete generation, even after failed startup.
            print("Waiting for a source change to retry the development stack.", flush=True)
            while not stopped and not watcher.changed():
                time.sleep(0.25)
        if not stopped:
            print("Source change detected; rebuilding Brain and restarting all services.", flush=True)
            watcher.reset()
    return 0


if __name__ == "__main__":
    try:
        sys.exit(run())
    except (OSError, ValueError) as error:
        print(f"Development startup failed: {error}", file=sys.stderr)
        sys.exit(1)
