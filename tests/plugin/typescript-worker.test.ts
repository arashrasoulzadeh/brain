import { mkdtemp, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawn } from "node:child_process";

const directory = await mkdtemp(join(tmpdir(), "brain-ts-worker-"));
const entrypoint = join(directory, "index.mjs");
const descriptor = join(directory, "execution.json");
await writeFile(
  entrypoint,
  "export function handle(request) { return { accepted: true, operation: request.params.operation }; }\n",
);
await writeFile(
  descriptor,
  JSON.stringify({
    entrypoint,
    pluginId: "sample",
    version: "1.0.0",
    projectId: "project",
    operations: ["health"],
  }),
);

const worker = spawn(process.execPath, ["scripts/typescript-worker.mjs", descriptor], {
  stdio: ["pipe", "pipe", "pipe"],
});
const output = new Promise<string>((resolve, reject) => {
  let value = "";
  worker.stdout.on("data", (chunk) => {
    value += chunk.toString();
    if (value.includes("\n")) resolve(value.trim());
  });
  worker.on("error", reject);
});
worker.stdin.write(
  `${JSON.stringify({ jsonrpc: "2.0", id: "one", params: { pluginId: "sample", version: "1.0.0", projectId: "project", operation: "health" } })}\n`,
);
const response = JSON.parse(await output) as {
  id: string;
  result: { accepted: boolean; operation: string };
};
if (response.id !== "one" || !response.result.accepted || response.result.operation !== "health") {
  throw new Error("typescript worker did not execute the plugin");
}
worker.kill();
