"""Check documented public HTTP commands and SDK exports against tracked source."""
from pathlib import Path
import re

def exact(label, actual, documented):
    missing = actual - documented
    stale = documented - actual
    if missing or stale:
        raise ValueError(f"{label}: missing={sorted(missing)} stale={sorted(stale)}")
    return len(actual)

def audit(root: Path) -> str:
    http = (root / "docs/http-api.md").read_text()
    routes = set()
    for pattern in re.findall(r'mux\.Handle(?:Func)?\("([^"]+)"', (root / "internal/app/app.go").read_text()):
        routes.add(pattern if " " in pattern else "ANY " + pattern)
    documented_routes = set(re.findall(r'^\| `((?:GET|POST|DELETE|ANY|PUT|PATCH) /[^`]+)` \|', http, re.M))
    route_count = exact("HTTP route documentation", routes, documented_routes)
    admin_source = (root / "internal/app/admin.go").read_text().split("func requiresAdminConfirmation")[0]
    operations = {op for line in re.findall(r'^\s*case ([^:\n]+):', admin_source, re.M) for op in re.findall(r'"([a-z][a-z.-]+)"', line)}
    admin_doc, function_doc = http.split("## Function commands", 1)
    documented_ops = set(re.findall(r'^\| `([a-z][a-z.-]+)` \|', admin_doc, re.M))
    admin_count = exact("Admin action documentation", operations, documented_ops)
    fn_source = (root / "internal/app/admin_functions.go").read_text()
    fn_ops = set(re.findall(r'command\.Operation [!=]= "(functions\.[a-z]+)"', fn_source))
    fn_count = exact("Function command documentation", fn_ops, set(re.findall(r'^\| `(functions\.[a-z]+)` \|', function_doc, re.M)))
    sdk = (root / "plugins/sdk-typescript/src/index.ts").read_text()
    exports = set(re.findall(r'^export (?:declare )?(?:async )?(?:type|interface|class|function|const|let|var|enum) (\w+)', sdk, re.M))
    for group in re.findall(r'^export (?:type )?\{([^}]+)\}', sdk, re.M):
        exports.update(name.strip().split(" as ")[-1] for name in group.split(",") if name.strip())
    sdk_doc = (root / "docs/sdk-reference.md").read_text()
    missing = exports - set(re.findall(r'`([A-Za-z_][A-Za-z0-9_]*)(?:<[^`>]+>)?`', sdk_doc))
    if missing:
        raise ValueError(f"SDK exports missing documentation: {sorted(missing)}")
    pages = root / "admin-nextjs/app"
    page_routes = {"/" if path.parent == pages else "/" + path.relative_to(pages).parent.as_posix() for path in pages.rglob("page.tsx")}
    page_doc = (root / "docs/admin-pages/README.md").read_text()
    page_count = exact("Admin page documentation", page_routes, set(re.findall(r'^\| \`(/[^\`]*)\` \|', page_doc, re.M)))
    manifests = sorted(root.glob("plugins/*/plugin.yaml")) + sorted(root.glob("examples/*/plugin.yaml"))
    expected_guides = {m.parent.name for m in manifests}
    guides = {p.stem for p in (root / "docs/plugin-guides").glob("*.md")} - {"index"}
    guide_count = exact("Plugin guides", expected_guides, guides)
    index = (root / "docs/plugin-guides/index.md").read_text()
    exact("Plugin guide index", expected_guides, set(re.findall(r'\]\(([a-z-]+)\.md\)', index)))
    lines = ["# Documentation coverage inventory", "", "Generated from source at build time. These are exact inventory checks, not a proof of semantic completeness or runtime test coverage.", "",
             "| Surface | Documented | Discovered |", "| --- | ---: | ---: |",
             f"| Mounted HTTP route patterns | {route_count} | {route_count} |",
             f"| Admin action commands | {admin_count} | {admin_count} |",
             f"| Function authoring commands | {fn_count} | {fn_count} |",
             f"| Public SDK entrypoint exports | {len(exports)} | {len(exports)} |",
             f"| Admin UI page patterns | {page_count} | {page_count} |",
             f"| Plugin guides | {guide_count} | {len(manifests)} |", "",
             "See [HTTP reference](../http-api.md), [SDK reference](../sdk-reference.md), [plugin guides](../plugin-guides/index.md) and [generated manifest reference](plugins/index.md).",
             "", "Client transport integration, deployment rehearsals, shell utilities, domain implementation details and narrative examples require human review beyond this inventory."]
    return "\n".join(lines) + "\n"
