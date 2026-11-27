import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("prepare_v8", ROOT / "scripts/prepare-v8-coverage.py")
PREPARE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(PREPARE)


class V8CoverageTest(unittest.TestCase):
    def test_only_virtual_source_map_keys_are_removed(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "raw"
            source.mkdir()
            profile = {"result": [{"url": "file:///repo/a.ts"}],
                       "source-map-cache": {"file:///repo/a.ts": {"data": {}},
                                            "node:internal/test": {}, "data:virtual": {}}}
            (source / "coverage.json").write_text(json.dumps(profile))
            PREPARE.prepare(source, root / "clean")
            output = json.loads((root / "clean/coverage.json").read_text())
            self.assertEqual(output["result"], profile["result"])
            self.assertEqual(output["source-map-cache"], {"file:///repo/a.ts": {"data": {}}})
            self.assertEqual(json.loads((source / "coverage.json").read_text()), profile)


if __name__ == "__main__":
    unittest.main()
