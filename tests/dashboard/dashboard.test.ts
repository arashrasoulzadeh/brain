import { adminRequest, FormAuditRecorder, FormAuditReader, authSettingsPage, browserQuickstart, createViewReplacementRegistry, fileManagementPage, hostOwnedAdminView, pluginHealthContribution, projectCreateRequest, projectOverviewContribution, recordBrowserPage, recordEditPage, signupSettingsPage, tableManagementPage } from "../../plugins/dashboard/index.js";
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
const result = createHostHandler(host, routes, { projectId: "operator", invocationId: "test-project-create" })(projectCreateRequest("dashboard", "0.1.0", "dev"));
if (typeof result !== "object") throw new Error("dashboard handler did not return a result");
if (calls.length !== 1) throw new Error("dashboard host was not invoked");
const readiness = adminRequest("dev", "dashboard", "0.1.0", "plugin.readiness", { targetPlugin: "auth", targetVersion: "0.1.0" });
if (!hostOwnedAdminView("dev").trusted) throw new Error("host-owned admin fixture was not trusted");
if (projectOverviewContribution("dev").view !== "project.overview" || pluginHealthContribution("dev").view !== "plugin.health") throw new Error("admin view contributions were not registered");
const views = createViewReplacementRegistry();
views.activate("dev", "project.overview", "custom-dashboard");
try { views.activate("dev", "project.overview", "second-dashboard"); throw new Error("duplicate view replacement accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "view replacement already active") throw error; }
try { views.submit("dev", "prod", "project.overview"); throw new Error("cross-project view submission accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "host project context mismatch") throw error; }
if (typeof createHostHandler(host, routes, { projectId: "dev", invocationId: "test-readiness" })(readiness) !== "object") throw new Error("readiness request was not dispatched");
try { createHostHandler(host, routes, { projectId: "dev", invocationId: "test-unknown" })(adminRequest("dev", "dashboard", "0.1.0", "plugin.unknown", {})); throw new Error("undeclared route was accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "operation not declared in plugin manifest") throw error; }
try { createHostHandler(host, routes, { projectId: "dev", invocationId: "test-cross-project" })(adminRequest("prod", "dashboard", "0.1.0", "plugin.enable", {})); throw new Error("cross-project host request accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "host project context mismatch") throw error; }
if (authSettingsPage("dev", { passwordEnabled: true, oidcProviders: ["google"], emailRecoveryEnabled: true }).title !== "Authentication settings") throw new Error("auth settings page failed");
if (signupSettingsPage("dev", { enabled: true, allowedDomains: ["example.com"], requireInvite: false }).title !== "Signup settings") throw new Error("signup settings page failed");
const browser = recordBrowserPage("dev", "users", [{ id: "1", name: "alice", secret: "x" }], ["id", "name"]);
if (browser.rows[0].secret !== undefined || browser.rows[0].name !== "alice") throw new Error("record browser exposure failed");
const edit = recordEditPage("dev", "users", "1", 2, { name: "bob", secret: "nope" }, ["name"]);
if (edit.patch.name !== "bob" || edit.patch.secret !== undefined || edit.expectedVersion !== 2) throw new Error("record edit validation failed");
if (tableManagementPage("dev", ["users", "orders"]).tables.length !== 2) throw new Error("table management page failed");
if (fileManagementPage("dev", [{ objectId: "avatar.png", size: 12, status: "published" }]).files[0].objectId !== "avatar.png") throw new Error("file management page failed");
const audits: { projectId: string; actorId: string; action: string; occurredAt: number }[] = [];
new FormAuditRecorder(audits).record("dev", "operator-1", "signup.update", 1000);
if (new FormAuditReader(audits).list("dev").length !== 1 || new FormAuditReader(audits).list("prod").length !== 0) throw new Error("form audit failed");
if (browserQuickstart("dev", "http://localhost:8080/api").steps.length !== 3) throw new Error("browser quickstart failed");
