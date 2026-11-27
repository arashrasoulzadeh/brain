import sys
from pathlib import Path
import unittest

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "scripts"))
from documentation_inventory import audit, exact

class InventoryTest(unittest.TestCase):
    def test_current_public_contracts_are_documented(self):
        result = audit(ROOT)
        self.assertIn("Documentation coverage inventory", result)
        self.assertIn("Public SDK entrypoint exports", result)

    def test_missing_and_stale_contracts_fail(self):
        self.assertEqual(exact("routes", {"a"}, {"a"}), 1)
        with self.assertRaisesRegex(ValueError, "missing=.*b"):
            exact("routes", {"a", "b"}, {"a"})
        with self.assertRaisesRegex(ValueError, "stale=.*b"):
            exact("routes", {"a"}, {"a", "b"})
