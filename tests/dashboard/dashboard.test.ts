import { adminRequest, handle, manifest, projectCreateRequest } from "../../plugins/dashboard/src/index.js";

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

const result = handle(projectCreateRequest("dev"));
if (typeof result !== "object") throw new Error("dashboard handler did not return a result");
