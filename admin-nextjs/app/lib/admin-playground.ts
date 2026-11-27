import type { ProjectHTTPRoute } from "./admin-types.js";

export function routeIdentity(route: ProjectHTTPRoute): string {
  return [route.plugin_id, route.version, route.operation].join(":");
}

export function searchRoutes(routes: readonly ProjectHTTPRoute[], query: string, method: string, auth: string): ProjectHTTPRoute[] {
  const needle = query.trim().toLowerCase();
  return routes.filter(route => (!method || route.method === method) && (!auth || route.auth === auth) &&
    [route.plugin_id, route.version, route.path, route.operation, route.capability, route.function].some(value => value.toLowerCase().includes(needle)));
}

export function preparePlayground(project: string, routes: readonly ProjectHTTPRoute[], identity: string, source: string) {
  if (!project) throw new Error("Select an active project first");
  if (!identity) throw new Error("Select a route first");
  const route = routes.find(candidate => routeIdentity(candidate) === identity);
  if (!route) throw new Error("Route no longer available; refresh and select it again");
  let payload: unknown;
  try { payload = JSON.parse(source || "{}"); } catch { throw new Error("Invalid JSON payload"); }
  const body = JSON.stringify({ operation: "plugin.invoke", project_id: project, plugin_id: route.plugin_id, version: route.version, target_operation: route.operation, payload });
  if (new TextEncoder().encode(body).byteLength > 64 * 1024) throw new Error("Admin request exceeds 64 KiB");
  return { identity, project, body };
}

export function decodePlaygroundResult(value: unknown): { success: boolean; message: string; output: string } {
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("Invalid host execution response");
  const result = value as Record<string, unknown>;
  const success = result.Success ?? result.success;
  if (typeof success !== "boolean") throw new Error("Missing worker execution status");
  const encoded = result.Payload ?? result.payload;
  let output: unknown = result;
  if (typeof encoded === "string" && encoded) {
    try {
      const decoded = new TextDecoder("utf-8", { fatal: true }).decode(Uint8Array.from(atob(encoded), character => character.charCodeAt(0)));
      try { output = JSON.parse(decoded); } catch { output = decoded; }
    } catch { output = encoded; }
  }
  return { success, message: success ? "Host execution completed" : "Plugin error: " + String(result.ErrorCode ?? result.errorCode ?? "execution failed"), output: JSON.stringify(output, null, 2) };
}
