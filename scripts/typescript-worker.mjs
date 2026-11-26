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
const routeHelpers = new Map([
  ["sdk.dispatch", { exported: "dispatchRoute", input: (request) => request.params.payload ?? null }],
  ["sdk.request", { exported: "dispatchRequest", input: (request) => request }],
]);
if (descriptor.routes.some((route) => {
  if (route.helper === "host.model") return false; // Executed exclusively by the Go registry.
  if (route.helper === "sdk.pipeline") return !validPipeline(route.pipeline);
  if (route.helper === "sdk.response") return !validResponse(route.response);
  if (route.helper === "sdk.context") return typeof plugin[route.function] !== "function";
  if (route.helper) {
    const helper = routeHelpers.get(route.helper);
    return !helper || typeof plugin[helper.exported] !== "function";
  }
  return typeof handler !== "function" && typeof plugin.functions !== "function" && typeof plugin[route.function] !== "function";
})) {
  process.stderr.write("plugin must export handle or createHandler\n");
  process.exit(65);
}

if (process.env.BRAIN_WORKER_READY === "1") process.stdout.write('{"ready":true}\n');

const input = readline.createInterface({ input: process.stdin, crlfDelay: Infinity });
for await (const line of input) {
  if (!line.trim()) continue;
  let request;
  let acceptingObservations = true;
  try {
    request = JSON.parse(line);
    if (JSON.stringify(request).length > 1_048_576) throw new Error("request too large");
    const params = request.params ?? {};
    if (params.pluginId !== descriptor.pluginId || params.version !== descriptor.version || params.projectId !== descriptor.projectId) {
      throw new Error("execution context mismatch");
    }
    const observations = [];
    const fire = (event, payload) => {
      if (!acceptingObservations || params.operation.startsWith("@observe/")) throw new Error("observation cannot be published outside a publisher invocation");
      if (observations.length >= 16) throw new Error("observation event limit exceeded");
      const spec = (descriptor.events ?? []).find((item) => item.name === event);
      if (!spec || !payload || typeof payload !== "object" || Array.isArray(payload)) throw new Error("undeclared or invalid observation");
      const serialized = JSON.stringify(payload);
      if (!serialized || Buffer.byteLength(serialized, "utf8") > 16 * 1024) throw new Error("observation payload too large");
      const snapshot = JSON.parse(serialized);
      if (!validEventPayload(spec, snapshot)) throw new Error("invalid observation fields");
      observations.push({ event, payload: snapshot });
    };
    const hostContext = params.hostContext;
    const available = new Set(Array.isArray(hostContext?.available) ? hostContext.available : []);
    const context = Object.freeze({
      project: Object.freeze({ id: descriptor.projectId, ...(hostContext?.projectVerified ? { active: true } : {}) }),
      plugin: Object.freeze({ id: descriptor.pluginId, version: descriptor.version }),
      execution: Object.freeze({ operation: params.operation, ...(hostContext?.invocationId ? { invocationId: hostContext.invocationId } : {}) }),
      ...(hostContext ? { config: Object.freeze({ ...hostContext.config }) } : {}),
      plugins: Object.freeze({ observe: Object.freeze({ fire }), has: (id) => typeof id === "string" && [...available].some((dependency) => dependency === id || dependency.split("@")[0] === id) }),
    });
    // Never accept a context supplied by the request or payload.
    request = { ...request, context };
    const route = descriptor.routes.find((candidate) => candidate.operation === params.operation);
    if (!route) throw new Error("operation not declared in plugin manifest");
    if (route.helper === "host.model") throw new Error("host-only model operation");
    const helper = routeHelpers.get(route.helper);
    const result = route.helper === "sdk.pipeline"
      ? await executePipeline(route.pipeline, params.payload ?? null, context)
      : route.helper === "sdk.response"
      ? mapResponse(route.response, params.payload, context)
      : route.helper === "sdk.context"
      ? await plugin[route.function](context, params.payload ?? null)
      : helper
      ? await plugin[helper.exported](route.function, helper.input(request), context)
      : typeof handler === "function" ? await handler(request, context) : await invokeDeclaredRoute(plugin, route.function, request);
    acceptingObservations = false;
    process.stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id: request.id, result: { protocol: params.protocol, success: true, payload: result, ...(observations.length ? { observations } : {}) } })}\n`);
  } catch (error) {
    acceptingObservations = false;
    process.stdout.write(`${JSON.stringify({ jsonrpc: "2.0", id: request?.id ?? "", errorCode: error instanceof Error ? error.message : "plugin execution failed" })}\n`);
  }
}

async function invokeDeclaredRoute(plugin, functionName, request) {
  const directExport = plugin[functionName];
  if (typeof directExport === "function") return directExport(request.params.payload ?? null);
  const functionMap = typeof plugin.functions === "function" ? plugin.functions() : null;
  const mappedExport = functionMap?.[functionName];
  if (functionMap && Object.hasOwn(functionMap, functionName) && typeof mappedExport === "function") return mappedExport(request, request.context);
  throw new Error(`declared plugin function is not exported: ${functionName}`);
}

function validEventPayload(spec, payload) {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) return false;
  const fields = Object.entries(spec.fields ?? {});
  return Object.keys(payload).length === fields.length && fields.every(([key, kind]) => {
    if (!Object.hasOwn(payload, key) || typeof payload[key] !== kind) return false;
    if (kind === "string") return Buffer.byteLength(payload[key], "utf8") <= 1024;
    return kind === "boolean" || (kind === "number" && Number.isFinite(payload[key]));
  });
}

function responseKey(key) {
  return typeof key === "string" && /^[A-Za-z_][A-Za-z0-9_]*$/.test(key) && !["__proto__", "constructor", "prototype"].includes(key);
}

function validResponse(spec) {
  if (!spec || (!Array.isArray(spec.input) && spec.input !== null) || !spec.output || typeof spec.output !== "object" || Array.isArray(spec.output)) return false;
  const fields = spec.input ?? [];
  const output = Object.entries(spec.output);
  if (fields.length > 32 || !output.length || output.length > 32) return false;
  const names = new Set();
  for (const field of fields) {
    if (!field || !responseKey(field.name) || names.has(field.name) || !["string", "value"].includes(field.type) ||
        (field.type !== "string" && (field.trim || field.default !== undefined)) ||
        (field.default !== undefined && typeof field.default !== "string") ||
        (field.trim !== undefined && typeof field.trim !== "boolean") ||
        (field.required !== undefined && typeof field.required !== "boolean") ||
        (field.required && (typeof field.error !== "string" || !field.error))) return false;
    names.add(field.name);
  }
  return output.every(([key, value]) => responseKey(key) &&
    (value === null || typeof value === "boolean" || (typeof value === "number" && Number.isFinite(value)) ||
      (typeof value === "string" && (!value.startsWith("$") || (value.startsWith("$input.") && names.has(value.slice(7))) || value === "$project.id" || value === "$plugin.id"))));
}

function mapResponse(spec, payload, context) {
  const values = new Map();
  for (const field of spec.input ?? []) {
    const raw = payload && typeof payload === "object" && Object.hasOwn(payload, field.name) ? payload[field.name] : undefined;
    let value = raw ?? null;
    if (field.type === "string") {
      value = typeof raw === "string" ? raw : field.default ?? "";
      if (field.trim) value = value.trim();
    }
    if (field.required && (value === null || value === "")) throw new Error(field.error);
    values.set(field.name, value);
  }
  return Object.fromEntries(Object.entries(spec.output).map(([key, value]) =>
    [key, value === "$project.id" ? context.project.id : value === "$plugin.id" ? context.plugin.id :
      typeof value === "string" && value.startsWith("$input.") ? values.get(value.slice(7)) : value]));
}

function pipelineHelpers() { return new Map([
  ["string.trim", (value) => { if (typeof value !== "string") throw new Error("expected string"); return value.trim(); }],
  ["string.nonempty", (value) => { if (typeof value !== "string" || !value.trim()) throw new Error("expected nonempty string"); return value; }],
  ["value.identity", (value) => value],
  ["number.integer", (value) => { if (!Number.isSafeInteger(value)) throw new Error("expected integer"); return value; }],
  ["boolean.require", (value) => { if (typeof value !== "boolean") throw new Error("expected boolean"); return value; }],
]); }

function pipelineReference(value, available) {
  if (value === "$input" || value === "$project.id" || value === "$plugin.id") return true;
  if (typeof value !== "string") return false;
  const parts = value.split(".");
  if (parts[0] === "$input") return parts.length <= 8 && parts.slice(1).every(responseKey);
  return parts[0] === "$steps" && parts.length >= 2 && parts.length <= 8 &&
    available.has(parts[1]) && parts.slice(1).every(responseKey);
}

function validPipeline(spec) {
  if (!spec || !Array.isArray(spec.steps) || spec.steps.length > 32) return false;
  const names = new Set();
  for (const step of spec.steps) {
    if (!step || !responseKey(step.save) || names.has(step.save) ||
        !(pipelineHelpers().has(step.helper) || step.helper === "observe.fire") || !pipelineReference(step.input, names)) return false;
    names.add(step.save);
  }
  return pipelineReference(spec.return, names);
}

async function executePipeline(spec, input, context) {
  const values = new Map();
  const helpers = pipelineHelpers();
  const resolve = (reference) => {
    if (reference === "$project.id") return context.project.id;
    if (reference === "$plugin.id") return context.plugin.id;
    const parts = reference.split(".");
    let value = parts[0] === "$input" ? input : values.get(parts[1]);
    for (const key of parts.slice(parts[0] === "$input" ? 1 : 2)) {
      if (!value || typeof value !== "object" || !Object.hasOwn(value, key)) throw new Error("missing pipeline value");
      value = value[key];
    }
    return value;
  };
  for (const step of spec.steps) {
    const value = resolve(step.input);
    if (step.helper === "observe.fire") {
      if (!value || typeof value !== "object" || !Object.hasOwn(value, "event") || !Object.hasOwn(value, "payload")) throw new Error("invalid observation input");
      context.plugins.observe.fire(value.event, value.payload);
      values.set(step.save, null);
    } else {
      values.set(step.save, await helpers.get(step.helper)(value));
    }
  }
  return resolve(spec.return);
}
