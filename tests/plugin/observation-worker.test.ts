import { deepEqual, equal } from "node:assert/strict";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import { test } from "node:test";

test("worker stages declared observations only on successful invocations", async () => {
  const directory = await mkdtemp(join(tmpdir(), "brain-observe-"));
  try {
    const entrypoint = join(directory, "plugin.mjs");
    const descriptor = join(directory, "descriptor.json");
    await writeFile(entrypoint, 'export function connect(ctx, input) { ctx.plugins.observe.fire(input.event, input.payload); if (input.fail) throw new Error("failed"); return { ok: true }; }');
    await writeFile(descriptor, JSON.stringify({
      entrypoint, projectId: "project", pluginId: "rooms", version: "1",
      routes: [
        { operation: "rooms.connect", function: "connect", helper: "sdk.context" },
        { operation: "@observe/rooms/connected", function: "connect", helper: "sdk.context" },
      ],
      events: [{ name: "connected", fields: { roomId: "string" } }],
    }));
    const params = { protocol: 1, projectId: "project", pluginId: "rooms", version: "1", operation: "rooms.connect" };
    const valid = { event: "connected", payload: { roomId: "one" } };
    const requests = [
      { ...params, payload: valid },
      { ...params, payload: { ...valid, fail: true } },
      { ...params, payload: { ...valid, event: "undeclared" } },
      { ...params, payload: { ...valid, payload: { roomId: "one", projectId: "forged" } } },
      { ...params, operation: "@observe/rooms/connected", payload: valid },
    ];
    const worker = spawnSync(process.execPath, ["scripts/typescript-worker.mjs", descriptor], { input: requests.map((request, index) => JSON.stringify({ id: index, params: request })).join("\n") + "\n", encoding: "utf8", timeout: 5000 });
    equal(worker.status, 0, worker.stderr);
    const results = worker.stdout.trim().split("\n").map((line) => JSON.parse(line) as { result?: { observations?: unknown }; errorCode?: string });
    deepEqual(results[0].result?.observations, [valid]);
    for (const result of results.slice(1)) {
      equal(result.result, undefined);
      equal(typeof result.errorCode, "string");
    }
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});
