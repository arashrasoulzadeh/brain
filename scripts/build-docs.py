#!/usr/bin/env python3
"""Stage repository documentation and generate plugin references with the host CLI."""
import argparse
import os
from pathlib import Path, PurePosixPath
import posixpath
import re
import shutil
import subprocess
from urllib.parse import quote, urlsplit

ROOT = Path(__file__).resolve().parents[1]
GITHUB = "https://github.com/arashrasoulzadeh/brain/blob/main/"
LINK = re.compile(r"(\]\()([^\s)]+)(\))")
FENCE = re.compile(r"^\s*(```+|~~~+)")
DOC_PREFIXES = ("docs/", "implementation/", "plugins/", "clients/", "protocol/", "tests/")
ROOT_DOCS = {"README.md", "CHANGELOG.md", "plan.md", "phases.md", "roadmap.md",
             "roadmap_backend.md", "roadmap_admin.md", "roadmap_client.md", "AGENTS.md"}


def source_map(root):
    paths = subprocess.check_output(["git", "ls-files", "-z"], cwd=root).decode().split("\0")
    result = {}
    for path in sorted(set(paths)):
        if not path.endswith(".md") or not (path in ROOT_DOCS or path.startswith(DOC_PREFIXES)):
            continue
        if not (root / path).is_file():
            raise ValueError(f"Missing tracked document: {path}")
        target = path.removeprefix("docs/") if path.startswith("docs/") else path
        if path == "README.md":
            target = "repository.md"
        if target in result.values():
            raise ValueError(f"Documentation destination collision: {target}")
        result[path] = target
    return result


def rewrite_links(content, source, target, mapping):
    def replace(match):
        raw = match.group(2)
        parsed = urlsplit(raw)
        if parsed.scheme or parsed.netloc or not parsed.path or parsed.path.startswith("/"):
            return match.group(0)
        candidate = posixpath.normpath(posixpath.join(posixpath.dirname(source), parsed.path))
        candidates = [candidate, candidate.rstrip("/") + "/README.md"]
        if candidate.endswith(".html"):
            candidates.insert(0, candidate[:-5] + ".md")
        destination = next((mapping[p] for p in candidates if p in mapping), None)
        if destination:
            link = posixpath.relpath(destination, posixpath.dirname(target) or ".")
        else:
            link = GITHUB + quote(candidate, safe="/")
        if parsed.query:
            link += "?" + parsed.query
        if parsed.fragment:
            link += "#" + parsed.fragment
        return match.group(1) + link + match.group(3)

    active_fence = None
    output = []
    for line in content.splitlines(keepends=True):
        fence = FENCE.match(line)
        if fence:
            marker = fence.group(1)
            if active_fence is None:
                active_fence = marker
            elif marker[0] == active_fence[0] and len(marker) >= len(active_fence):
                active_fence = None
            output.append(line)
        elif active_fence:
            output.append(line)
        else:
            # Do not rewrite links inside inline code examples.
            parts = re.split(r"(`+[^`]*`+)", line)
            output.append("".join(part if i % 2 else LINK.sub(replace, part)
                                  for i, part in enumerate(parts)))
    return "".join(output)


def generate(root, output, contract):
    mapping = source_map(root)
    if "docs/index.md" not in mapping:
        raise ValueError("Missing documentation homepage")
    output.mkdir(parents=True, exist_ok=True)
    for source, target in mapping.items():
        dest = output / target
        dest.parent.mkdir(parents=True, exist_ok=True)
        content = (root / source).read_text(encoding="utf-8")
        dest.write_text(rewrite_links(content, source, target, mapping), encoding="utf-8")
    reference = output / "reference" / "plugins"
    reference.mkdir(parents=True, exist_ok=True)
    manifests = sorted(root.glob("plugins/*/plugin.yaml")) + sorted(root.glob("examples/*/plugin.yaml"))
    if not manifests:
        raise ValueError("No plugin manifests found")
    catalog = ["# Plugin reference", "", "Generated from the host's validated manifest contract. "
               "The examples are not automatically installed.", ""]
    names = set()
    for manifest in manifests:
        relative = manifest.relative_to(root)
        name = "-".join(relative.parts[:-1])
        if name in names:
            raise ValueError(f"Duplicate plugin reference: {name}")
        names.add(name)
        rendered = subprocess.check_output([str(contract), "plugin-contract", str(manifest), "markdown"],
                                           cwd=root, text=True)
        if not rendered.strip():
            raise ValueError(f"Empty contract documentation: {relative}")
        raw = manifest.read_text(encoding="utf-8")
        page = rendered.rstrip() + "\n\n## Complete manifest\n\n"
        page += f"[Source manifest]({GITHUB}{relative.as_posix()})\n\n~~~yaml\n{raw.rstrip()}\n~~~\n"
        (reference / f"{name}.md").write_text(page, encoding="utf-8")
        catalog.append(f"- [{relative.parent.name}]({name}.md) — `{relative.as_posix()}`")
    catalog.extend(["", f"Manifest reference coverage: {len(names)}/{len(manifests)}.", "",
                    "This count measures manifest reference generation, not narrative or test coverage."])
    (reference / "index.md").write_text("\n".join(catalog) + "\n", encoding="utf-8")
    with (output / "index.md").open("a", encoding="utf-8") as home:
        home.write("\n## Explore Brain\n\n"
                   '<div class="grid cards" markdown>\n\n'
                   "- **Plugin API reference**\n\n    Browse generated operations, grants, models and events.\n\n"
                   "    [Explore plugins](reference/plugins/index.md)\n\n"
                   "- **Browser and mobile clients**\n\n    Integrate with project-scoped transports.\n\n"
                   "    [Browser](browser-quickstart.md) · [Flutter](flutter-quickstart.md)\n\n"
                   "- **Operations**\n\n    Deploy, recover and administer installations.\n\n"
                   "    [Runbooks](deployment-runbook.md) · [Admin pages](admin-pages/README.md)\n\n"
                   "- **Quality and roadmap**\n\n    Understand coverage, release checks and planned work.\n\n"
                   "    [Testing](tests/README.md) · [Roadmap](roadmap.md)\n\n"
                   "</div>\n\n[Changelog](CHANGELOG.md) · [Repository overview](repository.md)\n")
    shutil.copytree(root / "docs" / "assets", output / "assets", dirs_exist_ok=True)
    print(f"Staged {len(mapping)} Markdown sources; generated {len(names)}/{len(manifests)} manifest references.")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--contract", type=Path, required=True, help="Built brain executable")
    args = parser.parse_args()
    output = ROOT / ".generated-docs"
    if output.exists():
        shutil.rmtree(output)
    generate(ROOT, output, args.contract.resolve())


if __name__ == "__main__":
    main()
