import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).resolve().parents[2] / "scripts" / "check-go-coverage.py"


class GoCoverageCheckTest(unittest.TestCase):
    def run_profile(self, content):
        with tempfile.TemporaryDirectory() as directory:
            profile = Path(directory) / "go.out"
            profile.write_text(content, encoding="utf-8")
            return subprocess.run(
                [sys.executable, str(SCRIPT), str(profile)],
                capture_output=True, text=True, check=False,
            )

    def test_merges_covered_duplicate_blocks_in_either_order(self):
        for counts in [(0, 3), (3, 0)]:
            result = self.run_profile(
                "mode: atomic\n"
                f"brain/a.go:1.1,2.1 2 {counts[0]}\n"
                f"brain/a.go:1.1,2.1 2 {counts[1]}\n"
                "brain/a.go:3.1,3.2 0 0\n"
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("2/2 statements covered; 0 missing", result.stdout)

    def test_does_not_hide_uncovered_blocks_or_round_up(self):
        result = self.run_profile(
            "mode: atomic\n"
            "brain/a.go:1.1,2.1 10000 1\n"
            "brain/b.go:1.1,2.1 1 0\n"
            "brain/b.go:1.1,2.1 1 0\n"
        )
        self.assertEqual(result.returncode, 1)
        self.assertEqual(result.stdout.count("brain/b.go:1.1,2.1 1 0"), 1)
        self.assertIn("10000/10001 statements covered; 1 missing", result.stdout)

    def test_fails_closed_for_empty_malformed_or_inconsistent_profiles(self):
        for profile in [
            "", "mode: invalid\n", "mode: atomic\n", "mode: atomic\nmalformed\n",
            "mode: atomic\nbrain/a.go:1.1,2.1 0 0\n",
            "mode: atomic\nbrain/a.go:1.1,2.1 1 1\nbrain/a.go:1.1,2.1 2 0\n",
        ]:
            with self.subTest(profile=profile):
                result = self.run_profile(profile)
                self.assertEqual(result.returncode, 1)
                self.assertIn("Go coverage check failed:", result.stderr)


if __name__ == "__main__":
    unittest.main()
