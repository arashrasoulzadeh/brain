import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, mkdirSync, writeFileSync, readFileSync, rmSync, symlinkSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { test } from "node:test";

const root = process.cwd();
function fixture() {
  const cwd = mkdtempSync(join(tmpdir(), "brain-tooling-"));
  mkdirSync(join(cwd, "plugins", "nested"), { recursive: true });
  mkdirSync(join(cwd, "tests"));
  return cwd;
}
function run(cwd: string, script: string) {
  return spawnSync(process.execPath, [resolve(root, "scripts", script)], {
    cwd, encoding: "utf8", env: { ...process.env, npm_config_offline: "true" },
  });
}
test("lint accepts imports before declarations and ignores non-TypeScript files", () => {
  const cwd = fixture();
  try {
    writeFileSync(join(cwd, "plugins", "nested", "good.ts"), '// comment\n\nimport {\n  join,\n} from "node:path";\nconst result = join("a", "b");\n');
    writeFileSync(join(cwd, "tests", "good.ts"), "export const valid = true;\n");
    writeFileSync(join(cwd, "plugins", "ignored.txt"), "not TypeScript");
    const result = run(cwd, "lint-typescript.mjs");
    assert.equal(result.status, 0, result.stderr);
    assert.match(result.stdout, /passed for 2 files/);
  } finally { rmSync(cwd, { recursive: true, force: true }); }
});
test("lint reports forbidden types, logging, and imports below declarations", () => {
  const cwd = fixture();
  try {
    writeFileSync(join(cwd, "plugins", "nested", "bad.ts"), [
      "const value: " + "a" + "ny = 1;",
      'import { join } from "node:path";',
      "con" + "sole." + "log(value);",
    ].join("\n"));
    const result = run(cwd, "lint-typescript.mjs");
    assert.equal(result.status, 1);
    assert.match(result.stderr, /explicit .* is forbidden/);
    assert.match(result.stderr, /logging is forbidden/);
    assert.match(result.stderr, /bad\.ts:2: imports must be at the top/);
  } finally { rmSync(cwd, { recursive: true, force: true }); }
});
test("formatter formats actual files and propagates parse errors", () => {
  const cwd = fixture();
  try {
    symlinkSync(join(root, "node_modules"), join(cwd, "node_modules"), "junction");
    writeFileSync(join(cwd, "package.json"), '{"private":true}');
    const file = join(cwd, "plugins", "nested", "format.ts");
    writeFileSync(file, "const x={a:1,b:2}\n");
    writeFileSync(join(cwd, "tests", "fixture.ts"), "export const ok=true\n");
    const result = run(cwd, "format-typescript.mjs");
    assert.equal(result.status, 0, result.stderr);
    assert.equal(readFileSync(file, "utf8"), "const x = { a: 1, b: 2 };\n");
    writeFileSync(file, "const = ;");
    assert.notEqual(run(cwd, "format-typescript.mjs").status, 0);
  } finally { rmSync(cwd, { recursive: true, force: true }); }
});
