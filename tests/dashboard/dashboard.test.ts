import { DashboardService, adminRequest, createHandler, manifest, projectCreateRequest } from "../../plugins/dashboard/src/index.js";

const request = projectCreateRequest("dev");
if (manifest.id !== "dashboard" || request.params.operation !== "project.create") {
  throw new Error("dashboard TypeScript plugin contract failed");
}

const install = adminRequest("dev", "plugin.install", { targetPlugin: "auth", targetVersion: "1.0.0" });
const grant = adminRequest("dev", "plugin.grant", { targetPlugin: "auth", targetVersion: "1.0.0", capability: "auth.manage" });
const enable = adminRequest("dev", "plugin.enable", { targetPlugin: "auth", targetVersion: "1.0.0" });
if (install.params.operation !== "plugin.install" || grant.params.operation !== "plugin.grant" || enable.params.operation !== "plugin.enable") {
  throw new Error("generic dashboard operation contract failed");
}

const service = new DashboardService();
const result = createHandler(service)(projectCreateRequest("dev"));
if (typeof result !== "object") throw new Error("dashboard handler did not return a result");
service.createProject("project-a");
if (!service.hasProject("project-a")) throw new Error("dashboard project was not created");
try { service.createProject("project-a"); throw new Error("duplicate project was accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "project already exists") throw error; }
service.installPlugin("project-a", "auth", "0.1.0");
if (!service.isInstalled("project-a", "auth", "0.1.0")) throw new Error("dashboard plugin was not installed");
try { service.installPlugin("project-b", "auth", "0.1.0"); throw new Error("unknown project accepted installation"); } catch (error) { if (!(error instanceof Error) || error.message !== "unknown project") throw error; }
