import readline from "node:readline";
import { pathToFileURL } from "node:url";

const manifestPath = process.argv[2];
if (!manifestPath) {
  process.stderr.write("manifest path is required\n");
  process.exit(64);
}

const modulePath = pathToFileURL(manifestPath).href;
const plugin = await import(modulePath);
const handler = plugin.handle ?? plugin.createHandler?.();
if (typeof handler !== "function") {
  process.stderr.write("plugin must export handle or createHandler\n");
  process.exit(65);
}

const input = readline.createInterface({ input: process.stdin, crlfDelay: Infinity });
for await (const line of input) {
  if (!line.trim()) continue;
  let request;
  try {
    request = JSON.parse(line);
    if (JSON.stringify(request).length > 1_048_576) throw new Error("request too large");
    const result = await handler(request);
    process.stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id: request.id, result })}\n`);
  } catch (error) {
    process.stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id: request?.id ?? "", errorCode: error instanceof Error ? error.message : "plugin execution failed" })}\n`);
  }
}
