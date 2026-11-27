import { deepEqual, equal, throws } from "node:assert/strict";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawnSync } from "node:child_process";
import { test } from "node:test";
import { callHelper, helperNames } from "../../plugins/sdk-typescript/src/index.js";

test("shared helper registry and worker pipeline execute the same normalization", async () => {
  equal(callHelper("string.upper", callHelper("string.trim", " room ")), "ROOM");
  throws(() => callHelper("number.integer", 1.5));
  equal(helperNames().includes("string.upper"), true);
  const root = await mkdtemp(join(tmpdir(), "brain-pipeline-"));
  try {
    const entrypoint = join(root, "plugin.mjs");
    const descriptor = join(root, "execution.json");
    await writeFile(entrypoint, "export {};");
    const pipeline = { steps: [{ helper: "string.trim", input: "$input.name", save: "trimmed" }, { helper: "string.upper", input: "$steps.trimmed", save: "name" }], return: "$steps.name" };
    await writeFile(descriptor, JSON.stringify({ entrypoint, projectId: "dev", pluginId: "sample", version: "1", routes: [{ operation: "normalize", function: "normalize", helper: "sdk.pipeline", pipeline }] }));
    const params = { protocol: 1, projectId: "dev", pluginId: "sample", version: "1", operation: "normalize" };
    const result = spawnSync(process.execPath, ["scripts/typescript-worker.mjs", descriptor], { input: [JSON.stringify({ id: "one", params: { ...params, payload: { name: " room " } } }), JSON.stringify({ id: "two", params: { ...params, payload: { name: 7 } } })].join("\n") + "\n", encoding: "utf8", timeout: 5000 });
    equal(result.status, 0, result.stderr);
    const [success, failure] = result.stdout.trim().split("\n").map((line) => JSON.parse(line));
    deepEqual(success.result, { protocol: 1, success: true, payload: "ROOM" });
    equal(failure.errorCode, "expected string");
  } finally { await rm(root, { recursive: true, force: true }); }
});
