import { adminRequest, authSettingsPage, projectCreateRequest, signupSettingsPage } from "../../plugins/dashboard/index.js";
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
  { operation: "plugin.readiness", function: "showReadiness", capability: "plugin.readiness" },
] as const;
const calls: unknown[] = [];
const host = { invoke(projectId: string, operation: string, capability: string, payload: unknown) { calls.push({ projectId, operation, capability, payload }); return { acceptedByHost: true }; } };
const result = createHostHandler(host, routes)(projectCreateRequest("dashboard", "0.1.0", "dev"));
if (typeof result !== "object") throw new Error("dashboard handler did not return a result");
if (calls.length !== 1) throw new Error("dashboard host was not invoked");
const readiness = adminRequest("dev", "dashboard", "0.1.0", "plugin.readiness", { targetPlugin: "auth", targetVersion: "0.1.0" });
if (typeof createHostHandler(host, routes)(readiness) !== "object") throw new Error("readiness request was not dispatched");
try { createHostHandler(host, routes)(adminRequest("dev", "dashboard", "0.1.0", "plugin.unknown", {})); throw new Error("undeclared route was accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "operation not declared in plugin manifest") throw error; }
if (authSettingsPage("dev", { passwordEnabled: true, oidcProviders: ["google"], emailRecoveryEnabled: true }).title !== "Authentication settings") throw new Error("auth settings page failed");
if (signupSettingsPage("dev", { enabled: true, allowedDomains: ["example.com"], requireInvite: false }).title !== "Signup settings") throw new Error("signup settings page failed");
