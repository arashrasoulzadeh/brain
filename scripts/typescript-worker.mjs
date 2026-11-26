import readline from "node:readline";
import { readFile } from "node:fs/promises";
import { pathToFileURL } from "node:url";

const descriptorPath = process.argv[2];
if (!descriptorPath) {
  process.stderr.write("execution descriptor path is required\n");
  process.exit(64);
}

const descriptor = JSON.parse(await readFile(descriptorPath, "utf8"));
const declarative = descriptor.runtime === "declarative";
if ((declarative ? Boolean(descriptor.entrypoint) : !descriptor.entrypoint) || !descriptor.pluginId || !descriptor.version || !descriptor.projectId || !Array.isArray(descriptor.routes)) {
  process.stderr.write("invalid execution descriptor\n");
  process.exit(65);
}
const plugin = declarative ? Object.freeze({}) : await import(pathToFileURL(descriptor.entrypoint).href);
const handler = plugin.handle ?? plugin.createHandler?.();
const routeHelpers = new Map([
  ["sdk.dispatch", { exported: "dispatchRoute", input: (request) => request.params.payload ?? null }],
  ["sdk.request", { exported: "dispatchRequest", input: (request) => request }],
]);
if (descriptor.routes.some((route) => {
  if (route.helper === "sdk.response" || route.helper === "sdk.resource") return !validResponse(route.response);
  if (declarative) return true;
  if (route.helper) {
    const helper = routeHelpers.get(route.helper);
    return !helper || typeof plugin[helper.exported] !== "function";
  }
  return typeof handler !== "function" && typeof plugin.functions !== "function" && typeof plugin[route.function] !== "function";
})) {
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
    const helper = routeHelpers.get(route.helper);
    const result = (route.helper === "sdk.response" || route.helper === "sdk.resource")
      ? mapResponse(route.response, params.payload)
      : helper
      ? await plugin[helper.exported](route.function, helper.input(request))
      : typeof handler === "function" ? await handler(request) : await invokeDeclaredRoute(plugin, route.function, request);
    process.stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id: request.id, result: { protocol: params.protocol, success: true, payload: result } })}\n`);
  } catch (error) {
    process.stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id: request?.id ?? "", errorCode: error instanceof Error ? error.message : "plugin execution failed" })}\n`);
  }
}

async function invokeDeclaredRoute(plugin, functionName, request) {
  const directExport = plugin[functionName];
  if (typeof directExport === "function") return directExport(request.params.payload ?? null);
  const functionMap = typeof plugin.functions === "function" ? plugin.functions() : null;
  const mappedExport = functionMap?.[functionName];
  if (functionMap && Object.hasOwn(functionMap, functionName) && typeof mappedExport === "function") return mappedExport(request);
  throw new Error(`declared plugin function is not exported: ${functionName}`);
}

function responseKey(key) {
  return typeof key === "string" && /^[A-Za-z_][A-Za-z0-9_]*$/.test(key) && !["__proto__", "constructor", "prototype"].includes(key);
}

function validResponse(spec) {
  if (!spec || (!Array.isArray(spec.input) && spec.input !== null) ||
      (spec.output !== undefined && (!spec.output || typeof spec.output !== "object" || Array.isArray(spec.output)))) return false;
  const fields = spec.input ?? [];
  const output = Object.entries(spec.output ?? {});
  if (fields.length > 32 || output.length > 32 || (output.length === 0) === !spec.result || (spec.merge && spec.result)) return false;
  const names = new Map();
  for (const field of fields) {
    if (!field || !responseKey(field.name) || names.has(field.name) || !["string", "value", "object", "array", "boolean", "number"].includes(field.type) ||
        (field.type !== "string" && (field.trim || field.default !== undefined)) ||
        (field.default !== undefined && typeof field.default !== "string") ||
        (field.trim !== undefined && typeof field.trim !== "boolean") ||
        (field.required !== undefined && typeof field.required !== "boolean") ||
        (field.required && (typeof field.error !== "string" || !field.error))) return false;
    names.set(field.name, field.type);
  }
  if ((spec.result && !validReference(spec.result, names)) || (spec.merge && names.get(spec.merge) !== "object")) return false;
  return output.every(([key, value]) => responseKey(key) &&
    (value === null || typeof value === "boolean" || (typeof value === "number" && Number.isFinite(value)) ||
      (typeof value === "string" && (!value.startsWith("$") || validReference(value, names)))));
}

function validReference(value, fields) {
  if (typeof value !== "string") return false;
  if (value.startsWith("$input.")) return fields.has(value.slice(7));
  if (value.startsWith("$length.")) return ["string", "array"].includes(fields.get(value.slice(8)));
  return false;
}

function readReference(value, values) {
  if (typeof value !== "string") return value;
  if (value.startsWith("$input.")) return values.get(value.slice(7));
  if (value.startsWith("$length.")) return values.get(value.slice(8))?.length ?? 0;
  return value;
}

function mapResponse(spec, payload) {
  const values = new Map();
  for (const field of spec.input ?? []) {
    const raw = payload && typeof payload === "object" && Object.hasOwn(payload, field.name) ? payload[field.name] : undefined;
    let value = raw ?? null;
    if (field.type === "string") {
      value = typeof raw === "string" ? raw : field.default ?? "";
      if (field.trim) value = value.trim();
    } else if (field.type === "object" && (value === null || typeof value !== "object" || Array.isArray(value))) {
      value = null;
    } else if (field.type === "array" && !Array.isArray(value)) {
      value = null;
    } else if (field.type === "boolean" && typeof value !== "boolean") {
      value = null;
    } else if (field.type === "number" && (typeof value !== "number" || !Number.isFinite(value))) {
      value = null;
    }
    if (field.required && (value === null || value === "")) throw new Error(field.error);
    values.set(field.name, value);
  }
  if (spec.result) return readReference(spec.result, values);
  return Object.fromEntries([
    ...Object.entries(spec.merge ? values.get(spec.merge) ?? {} : {}),
    ...Object.entries(spec.output).map(([key, value]) => [key, readReference(value, values)]),
  ]);
}
