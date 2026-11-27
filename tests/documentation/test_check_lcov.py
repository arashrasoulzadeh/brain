import importlib.util
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("check_lcov", ROOT / "scripts/check-lcov.py")
CHECK = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(CHECK)


class CoverageTest(unittest.TestCase):
    def test_missing_and_uncovered_files_fail(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            one = root / "one.dart"
            two = root / "two.dart"
            one.touch()
            two.touch()
            report = root / "lcov.info"
            report.write_text(f"SF:{one}\nDA:1,0\nend_of_record\n")
            self.assertEqual(CHECK.check(report, root), [f"{one}:1", f"{two}: no measured lines"])

    def test_duplicate_records_accumulate_counts(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            one = root / "one.dart"
            one.touch()
            report = root / "lcov.info"
            report.write_text(f"SF:{one}\nDA:1,0\nend_of_record\nSF:{one}\nDA:1,2\nend_of_record\n")
            self.assertEqual(CHECK.check(report, root), [])


if __name__ == "__main__":
    unittest.main()
