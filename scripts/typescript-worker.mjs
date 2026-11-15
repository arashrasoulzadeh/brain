import readline from "node:readline";
import { readFile } from "node:fs/promises";
import { pathToFileURL } from "node:url";

const descriptorPath = process.argv[2];
if (!descriptorPath) {
  process.stderr.write("execution descriptor path is required\n");
  process.exit(64);
}

const descriptor = JSON.parse(await readFile(descriptorPath, "utf8"));
if (!descriptor.entrypoint || !descriptor.pluginId || !descriptor.version || !descriptor.projectId || !Array.isArray(descriptor.routes)) {
  process.stderr.write("invalid execution descriptor\n");
  process.exit(65);
}
const modulePath = pathToFileURL(descriptor.entrypoint).href;
const plugin = await import(modulePath);
const handler = plugin.handle ?? plugin.createHandler?.();
if (typeof handler !== "function" && typeof plugin.functions !== "function" && descriptor.routes.some((route) => typeof plugin[route.function] !== "function")) {
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
    const params = request.params ?? {};
    if (params.pluginId !== descriptor.pluginId || params.version !== descriptor.version || params.projectId !== descriptor.projectId) {
      throw new Error("execution context mismatch");
    }
    const route = descriptor.routes.find((candidate) => candidate.operation === params.operation);
    if (!route) throw new Error("operation not declared in plugin manifest");
    const result = typeof handler === "function" ? await handler(request) : await invokeDeclaredRoute(plugin, route.function, request);
    process.stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id: request.id, result })}\n`);
  } catch (error) {
    process.stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id: request?.id ?? "", errorCode: error instanceof Error ? error.message : "plugin execution failed" })}\n`);
  }
}

async function invokeDeclaredRoute(plugin, functionName, request) {
  const directExport = plugin[functionName];
  if (typeof directExport === "function") return directExport(request.params.payload ?? null);
  const functionMap = typeof plugin.functions === "function" ? plugin.functions() : null;
  const mappedExport = functionMap?.[functionName];
  if (typeof mappedExport === "function") return mappedExport(request);
  throw new Error(`declared plugin function is not exported: ${functionName}`);
}
