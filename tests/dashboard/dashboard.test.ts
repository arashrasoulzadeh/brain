import { adminRequest, projectCreateRequest } from "../../plugins/dashboard/index.js";
import { createHostHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = projectCreateRequest("dashboard", "0.1.0", "dev");
if (request.params.pluginId !== "dashboard" || request.params.operation !== "project.create") {
  throw new Error("dashboard TypeScript plugin contract failed");
}

const install = adminRequest("dev", "dashboard", "0.1.0", "plugin.install", { targetPlugin: "auth", targetVersion: "1.0.0" });
const grant = adminRequest("dev", "dashboard", "0.1.0", "plugin.grant", { targetPlugin: "auth", targetVersion: "1.0.0", capability: "auth.manage" });
const enable = adminRequest("dev", "dashboard", "0.1.0", "plugin.enable", { targetPlugin: "auth", targetVersion: "1.0.0" });
if (install.params.operation !== "plugin.install" || grant.params.operation !== "plugin.grant" || enable.params.operation !== "plugin.enable") {
  throw new Error("generic dashboard operation contract failed");
}

const routes = [
  { operation: "project.create", function: "createProject", capability: "project.manage" },
  { operation: "plugin.install", function: "installPlugin", capability: "plugin.install" },
  { operation: "plugin.grant", function: "grantCapability", capability: "plugin.grant" },
  { operation: "plugin.enable", function: "enablePlugin", capability: "plugin.enable" },
] as const;
const calls: unknown[] = [];
const host = { invoke(projectId: string, operation: string, capability: string, payload: unknown) { calls.push({ projectId, operation, capability, payload }); return { acceptedByHost: true }; } };
const result = createHostHandler(host, routes)(projectCreateRequest("dashboard", "0.1.0", "dev"));
if (typeof result !== "object") throw new Error("dashboard handler did not return a result");
if (calls.length !== 1) throw new Error("dashboard host was not invoked");
try { createHostHandler(host, routes)(adminRequest("dev", "dashboard", "0.1.0", "plugin.unknown", {})); throw new Error("undeclared route was accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "operation not declared in plugin manifest") throw error; }
