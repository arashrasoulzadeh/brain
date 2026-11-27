#!/usr/bin/env python3
"""Check exact line coverage and ensure every scoped source appears in LCOV."""
import argparse
from pathlib import Path


def check(report, source_root):
    root = source_root.resolve()
    files = {path.resolve(): {} for path in root.rglob("*.dart")}
    if not files:
        raise ValueError(f"No Dart sources in {root}")
    seen = set()
    current = None
    for line in report.read_text(encoding="utf-8").splitlines():
        if line.startswith("SF:"):
            current = Path(line[3:]).resolve()
            if current in files:
                seen.add(current)
        elif line.startswith("DA:") and current in files:
            number, count, *_ = line[3:].split(",")
            files[current][number] = files[current].get(number, 0) + int(count)
        elif line == "end_of_record":
            current = None
    missed = []
    for path, lines in sorted(files.items()):
        if path not in seen or not lines:
            missed.append(f"{path}: no measured lines")
        else:
            missed.extend(f"{path}:{line}" for line, count in sorted(lines.items()) if count == 0)
    return missed


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("report", type=Path)
    parser.add_argument("source_root", type=Path)
    args = parser.parse_args()
    missed = check(args.report, args.source_root)
    if missed:
        print("\n".join(missed))
        raise SystemExit(1)
    print("Every measured Dart source line is covered.")


if __name__ == "__main__":
    main()
