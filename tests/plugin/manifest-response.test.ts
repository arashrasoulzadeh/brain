import { deepEqual, equal } from "node:assert/strict";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";

const root = await mkdtemp(join(tmpdir(), "brain-response-"));
try {
  const entrypoint = join(root, "plugin.mjs");
  const descriptor = join(root, "execution.json");
  await writeFile(entrypoint, "export const unused = true;");
  const route = {
    operation: "sample.read", function: "read", helper: "sdk.response",
    response: {
      input: [
        { name: "name", type: "string", trim: true, required: true, error: "name required" },
        { name: "data", type: "value" },
      ],
      output: { name: "$input.name", data: "$input.data", ok: true },
    },
  };
  const config = { entrypoint, pluginId: "sample", version: "1", projectId: "p", routes: [route] };
  await writeFile(descriptor, JSON.stringify(config));
  const params = { protocol: 1, projectId: "p", pluginId: "sample", version: "1", operation: "sample.read" };
  const lines = [
    { ...params, payload: { name: " x ", data: false } },
    { ...params, payload: { name: "y" } },
    { ...params, payload: { name: " " } },
    { ...params, projectId: "other", payload: { name: "x" } },
  ].map((value, i) => JSON.stringify({ id: String(i), params: value })).join("\n") + "\n";
  const run = () => spawnSync(process.execPath, ["scripts/typescript-worker.mjs", descriptor], { input: lines, encoding: "utf8", timeout: 5000 });
  const result = run();
  equal(result.status, 0, result.stderr);
  const [first, second, invalid, denied] = result.stdout.trim().split("\n").map(line => JSON.parse(line) as { result?: unknown; errorCode?: string });
  deepEqual(first.result, { protocol: 1, success: true, payload: { name: "x", data: false, ok: true } });
  deepEqual(second.result, { protocol: 1, success: true, payload: { name: "y", data: null, ok: true } });
  equal(invalid.errorCode, "name required");
  equal(denied.errorCode, "execution context mismatch");
  await writeFile(descriptor, JSON.stringify({ ...config, routes: [{ ...route, response: { ...route.response, output: { name: "$input.unknown" } } }] }));
  equal(run().status, 65);
} finally {
  await rm(root, { recursive: true, force: true });
}
