import importlib.util
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("merge_go_coverage", ROOT / "scripts/merge-go-coverage.py")
MERGER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MERGER)


class GoCoverageTest(unittest.TestCase):
    def profile(self, root, name, content):
        path = root / name
        path.write_text(content, encoding="utf-8")
        return path

    def test_duplicates_are_merged_without_hiding_uncovered_blocks(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            unit = self.profile(root, "unit", "mode: atomic\na.go:1.1,2.1 2 0\na.go:2.1,3.1 3 0\n")
            cli = self.profile(root, "cli", "mode: atomic\na.go:1.1,2.1 2 4\nb.go:1.1,2.1 1 1\n")
            self.assertEqual(MERGER.merge([unit, cli]), {
                "a.go:1.1,2.1": (2, 4), "a.go:2.1,3.1": (3, 0), "b.go:1.1,2.1": (1, 1)
            })

    def test_empty_mismatched_and_invalid_profiles_fail(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            valid = self.profile(root, "valid", "mode: atomic\na.go:1.1,2.1 2 1\n")
            for content in ("mode: set\na.go:1.1,2.1 2 1\n", "mode: atomic\na.go:1.1,2.1 3 1\n", "mode: atomic\na.go:1.1,2.1 2 -1\n"):
                with self.assertRaises(ValueError):
                    MERGER.merge([valid, self.profile(root, "invalid", content)])
            with self.assertRaises(ValueError):
                MERGER.merge([self.profile(root, "empty", "mode: atomic\n")])
