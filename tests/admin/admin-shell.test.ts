import { readFile, readdir } from "node:fs/promises";
import { join } from "node:path";
import { test } from "node:test";
import assert from "node:assert/strict";

const root = process.cwd();
const pageFiles = [
  "admin-nextjs/app/view.tsx",
  "admin-nextjs/app/components/legacy-admin-page.tsx",
  "admin-nextjs/app/functions/[name]/view.tsx",
  "admin-nextjs/app/functions/new/view.tsx",
  "admin-nextjs/app/functions/view.tsx",
  "admin-nextjs/app/plugins/[plugin]/view.tsx",
  "admin-nextjs/app/workers/view.tsx",
  "admin-nextjs/app/extensions/[project]/[plugin]/[version]/[page]/view.tsx",
];

test("every admin page uses the one shared shell or its shared main-page alias", async () => {
  for (const relativePath of pageFiles) {
    const source = await readFile(join(root, relativePath), "utf8");
    const isMainPage = relativePath === "admin-nextjs/app/view.tsx";
    assert.ok(
      isMainPage || source.includes("AdminPageFrame"),
      `${relativePath} bypasses AdminPageFrame`,
    );
    assert.equal(
      source.match(/<AdminSidebar\b/g)?.length ?? 0,
      isMainPage ? 1 : 0,
      `${relativePath} declares a duplicate sidebar`,
    );
  }
});

test("the shared shell owns the only sidebar implementation", async () => {
  const source = await readFile(join(root, "admin-nextjs/app/components/admin-chrome.tsx"), "utf8");
  assert.equal(source.match(/export function AdminSidebar\b/g)?.length ?? 0, 1);
  assert.equal(source.match(/<aside className="sidebar">/g)?.length ?? 0, 1);
  const menu = await readFile(join(root, "admin-nextjs/app/lib/admin-menu.ts"), "utf8");
  assert.ok(menu.includes('"/audit"'), "shared sidebar discovery must include Audit");
});

test("authenticated route content is mounted below the persistent root shell", async () => {
  const layout = await readFile(join(root, "admin-nextjs/app/layout.tsx"), "utf8");
  const shell = await readFile(join(root, "admin-nextjs/app/components/admin-shell.tsx"), "utf8");
  const frame = await readFile(join(root, "admin-nextjs/app/components/admin-chrome.tsx"), "utf8");
  assert.match(layout, /<AdminShell>\{children\}<\/AdminShell>/);
  assert.match(shell, /export function AdminShell/);
  assert.match(shell, /!authenticated/);
  assert.doesNotMatch(frame, /return <main className="shell">/);
});

test("all mounted admin pages are inside project UI or system UI namespaces", async () => {
  const app = join(root, "admin-nextjs/app");
  const walk = async (directory: string): Promise<string[]> => {
    const entries = await readdir(directory, { withFileTypes: true });
    const nested = await Promise.all(entries.map(entry => entry.isDirectory() ? walk(join(directory, entry.name)) : entry.name === "page.tsx" ? [join(directory, entry.name)] : []));
    return nested.flat();
  };
  const pages = await walk(app);
  assert.ok(pages.length >= 15);
  for (const page of pages) {
    const route = page.slice(app.length + 1);
    assert.ok(route.startsWith("[project]/ui/") || route.startsWith("%5Fsystem/ui/"), route);
  }
  assert.ok(pages.includes(join(app, "[project]/ui/functions/[name]/page.tsx")));
  assert.ok(pages.includes(join(app, "%5Fsystem/ui/page.tsx")));
});
