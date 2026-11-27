import { createRequire } from "node:module";
import { mkdirSync, openSync, closeSync } from "node:fs";
import { resolve } from "node:path";
import { spawn } from "node:child_process";
import { setTimeout as delay } from "node:timers/promises";
import assert from "node:assert/strict";

const require = createRequire(resolve(process.env.BRAIN_TEST_TOOLS ?? "tests/browser", "package.json"));
export const { chromium } = require("playwright");
export const artifacts = resolve(process.env.BRAIN_BROWSER_ARTIFACTS ?? "coverage/browser");
mkdirSync(artifacts, { recursive: true });

export function start(command, args, cwd, name) {
  const descriptor = openSync(resolve(artifacts, name + ".log"), "w");
  const child = spawn(command, args, { cwd, detached: true, stdio: ["ignore", descriptor, descriptor], env: { ...process.env, NEXT_TELEMETRY_DISABLED: "1" } });
  closeSync(descriptor);
  let launchError;
  child.on("error", error => { launchError = error; });
  return {
    async ready(url, timeout = 120000) {
      const deadline = Date.now() + timeout;
      while (Date.now() < deadline) {
        if (launchError) throw launchError;
        if (child.exitCode !== null) throw new Error(name + " exited: " + child.exitCode);
        try {
          const response = await fetch(url, { signal: AbortSignal.timeout(10000) });
          if (response.ok) return;
        } catch { /* server may still be starting */ }
        await delay(500);
      }
      throw new Error(name + " did not become ready: " + url);
    },
    close() {
      if (child.pid && child.exitCode === null) {
        try { process.kill(-child.pid, "SIGTERM"); } catch (error) { if (error.code !== "ESRCH") throw error; }
      }
    },
  };
}

export async function visible(locator) {
  await locator.waitFor({ state: "visible", timeout: 15000 });
  assert.equal(await locator.isVisible(), true);
}

export async function waitFor(check, description, timeout = 15000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    if (await check()) return;
    await delay(100);
  }
  assert.fail(description);
}
