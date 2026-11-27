#!/usr/bin/env python3
"""Preserve file-backed source maps and remove unsupported virtual cache entries."""
import argparse
import json
from pathlib import Path
from urllib.parse import urlsplit


def prepare(source, destination):
    destination.mkdir(parents=True, exist_ok=True)
    profiles = sorted(source.glob("*.json"))
    if not profiles:
        raise ValueError("No V8 coverage profiles")
    for profile in profiles:
        data = json.loads(profile.read_text(encoding="utf-8"))
        # c8 normalizes these keys with fileURLToPath. V8 also records internal
        # and virtual URLs that are not repository files and cannot be mapped.
        if "source-map-cache" in data:
            data["source-map-cache"] = {
                url: entry for url, entry in data["source-map-cache"].items()
                if urlsplit(url).scheme == "file"
            }
        (destination / profile.name).write_text(json.dumps(data), encoding="utf-8")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    parser.add_argument("destination", type=Path)
    args = parser.parse_args()
    prepare(args.source, args.destination)


if __name__ == "__main__":
    main()
