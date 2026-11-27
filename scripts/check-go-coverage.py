"""Require every distinct executable Go coverprofile block to have run."""

import argparse
import re
import sys
from pathlib import Path


BLOCK = re.compile(r"^(.+:\d+\.\d+,\d+\.\d+)\s+(\d+)\s+(\d+)$")


def check(path: Path) -> int:
    lines = path.read_text(encoding="utf-8").splitlines()
    if not lines or lines[0] not in {"mode: set", "mode: count", "mode: atomic"}:
        raise ValueError("missing or invalid Go coverage mode")
    blocks = {}
    for number, line in enumerate(lines[1:], 2):
        match = BLOCK.fullmatch(line)
        if match is None:
            raise ValueError(f"invalid coverage block on line {number}")
        location, raw_statements, raw_count = match.groups()
        statements, count = int(raw_statements), int(raw_count)
        previous = blocks.get(location)
        if previous is not None and previous[0] != statements:
            raise ValueError(f"inconsistent statement count for {location}")
        # Match Go's profile semantics: a block is covered if any run covered it.
        blocks[location] = (statements, count > 0 or bool(previous and previous[1]))
    total = sum(statements for statements, _ in blocks.values())
    if not total:
        raise ValueError("no executable statements in coverage profile")
    missed = 0
    for location, (statements, covered) in sorted(blocks.items()):
        if statements and not covered:
            print(f"{location} {statements} 0")
            missed += statements
    print(f"Go coverage: {total - missed}/{total} statements covered; {missed} missing")
    return int(missed > 0)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("profile", type=Path)
    args = parser.parse_args()
    try:
        return check(args.profile)
    except (OSError, ValueError) as error:
        print(f"Go coverage check failed: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
