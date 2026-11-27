#!/usr/bin/env python3
"""Merge Go coverage profiles by source block, retaining all uncovered statements."""
import argparse
from pathlib import Path


def merge(paths):
    blocks = {}
    for path in paths:
        lines = path.read_text(encoding="utf-8").splitlines()
        if not lines or lines[0] != "mode: atomic":
            raise ValueError(f"{path}: expected atomic coverage profile")
        for line in lines[1:]:
            location, statements, count = line.rsplit(" ", 2)
            statements, count = int(statements), int(count)
            if statements < 0 or count < 0:
                raise ValueError(f"{path}: invalid coverage counts")
            previous = blocks.get(location, (statements, 0))
            if previous[0] != statements:
                raise ValueError(f"{path}: inconsistent statement count for {location}")
            blocks[location] = (statements, previous[1] + count)
    if not blocks or not sum(statements for statements, _ in blocks.values()):
        raise ValueError("No executable Go coverage blocks")
    return blocks


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--require-complete", action="store_true")
    parser.add_argument("profiles", type=Path, nargs="+")
    args = parser.parse_args()
    blocks = merge(args.profiles)
    args.output.write_text("mode: atomic\n" + "".join(
        f"{location} {statements} {count}\n"
        for location, (statements, count) in sorted(blocks.items())
    ), encoding="utf-8")
    missed = [(location, statements) for location, (statements, count) in blocks.items() if statements and not count]
    total = sum(statements for statements, _ in blocks.values())
    uncovered = sum(statements for _, statements in missed)
    print(f"Go statements: {total - uncovered}/{total} ({100 * (total - uncovered) / total:.2f}%)")
    if args.require_complete and missed:
        for location, statements in sorted(missed):
            print(f"{location}: {statements} uncovered statements")
        raise SystemExit(1)


if __name__ == "__main__":
    main()
