import importlib.util
from pathlib import Path
import unittest
import tempfile

ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("build_docs", ROOT / "scripts/build-docs.py")
BUILD = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(BUILD)


class DocumentationTest(unittest.TestCase):
    def test_links_follow_relocated_sources(self):
        mapping = {"docs/index.md": "index.md", "docs/developers.md": "developers.md",
                   "README.md": "repository.md", "docs/admin-pages/README.md": "admin-pages/README.md"}
        source = "[guide](developers.md#setup) [repo](../README.md) [admin](admin-pages/)\n"
        expected = "[guide](developers.md#setup) [repo](repository.md) [admin](admin-pages/README.md)\n"
        self.assertEqual(BUILD.rewrite_links(source, "docs/index.md", "index.md", mapping), expected)

    def test_external_source_and_code_links(self):
        source = "[code](../cmd/brain/main.go) [web](https://example.com) [anchor](#a)\n"
        source += "```md\n[example](../README.md)\n```\n`[example](../README.md)`\n"
        expected = source.replace("../cmd/brain/main.go", BUILD.GITHUB + "cmd/brain/main.go")
        self.assertEqual(BUILD.rewrite_links(source, "docs/index.md", "index.md", {}), expected)

    def test_all_tracked_docs_have_unique_destinations(self):
        mapping = BUILD.source_map(ROOT)
        self.assertEqual(len(mapping), len(set(mapping.values())))
        for source in ("docs/index.md", "docs/admins.md", "docs/developers.md", "tests/README.md"):
            self.assertIn(source, mapping)

    def test_host_http_surface_is_documented(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            api = output / "http-api.md"
            api.write_text((ROOT / "docs/http-api.md").read_text(encoding="utf-8"), encoding="utf-8")
            routes, commands = BUILD.validate_http_docs(ROOT, output)
            self.assertGreater(routes, 0)
            self.assertGreater(commands, 0)
            api.write_text("", encoding="utf-8")
            with self.assertRaises(ValueError):
                BUILD.validate_http_docs(ROOT, output)

    def test_missing_or_empty_plugin_guide_sections_fail(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            path = output / "plugin-guides" / "plugins-test.md"
            path.parent.mkdir()
            with self.assertRaises(ValueError):
                BUILD.validate_plugin_guide(output, "plugins-test")
            for content in ("# Test\n## Developer\nnotes\n", "# Test\n## Developer\n## Administrator\nnotes\n"):
                path.write_text(content, encoding="utf-8")
                with self.assertRaises(ValueError):
                    BUILD.validate_plugin_guide(output, "plugins-test")
            path.write_text("# Test\n## Developer\nexample\n## Administrator\nrunbook\n", encoding="utf-8")
            self.assertEqual(BUILD.validate_plugin_guide(output, "plugins-test"), path)


if __name__ == "__main__":
    unittest.main()
