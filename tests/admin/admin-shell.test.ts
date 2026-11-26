import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { test } from "node:test";
import assert from "node:assert/strict";

const root = process.cwd();
const pageFiles = [
  "admin-nextjs/app/[legacy]/page.tsx",
  "admin-nextjs/app/audit/page.tsx",
  "admin-nextjs/app/database/page.tsx",
  "admin-nextjs/app/files/page.tsx",
  "admin-nextjs/app/functions/[name]/page.tsx",
  "admin-nextjs/app/functions/new/page.tsx",
  "admin-nextjs/app/functions/page.tsx",
  "admin-nextjs/app/live-sessions/page.tsx",
  "admin-nextjs/app/monitoring/page.tsx",
  "admin-nextjs/app/page.tsx",
  "admin-nextjs/app/plugins/[plugin]/page.tsx",
  "admin-nextjs/app/plugins/page.tsx",
  "admin-nextjs/app/projects/page.tsx",
  "admin-nextjs/app/routes/page.tsx",
  "admin-nextjs/app/system/page.tsx",
  "admin-nextjs/app/workers/page.tsx",
  "admin-nextjs/app/workflows/page.tsx",
];

test("every admin page uses the one shared shell or its shared main-page alias", async () => {
  for (const relativePath of pageFiles) {
    const source = await readFile(join(root, relativePath), "utf8");
    const isMainPage = relativePath === "admin-nextjs/app/page.tsx";
    const isAlias = source.includes('export { default } from "../page"') || source.includes('export { default } from "../components/legacy-admin-page"');
    assert.ok(
      isMainPage || isAlias || source.includes("AdminPageFrame"),
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
  assert.ok(source.includes('"audit"'), "shared sidebar discovery must include Audit");
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
