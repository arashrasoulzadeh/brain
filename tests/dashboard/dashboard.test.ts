import { handle, manifest, pluginEnableRequest, pluginGrantRequest, pluginInstallRequest, projectCreateRequest } from "../../plugins/dashboard/src/index.js";

const request = projectCreateRequest("dev");
if (manifest.id !== "dashboard" || request.params.operation !== "project.create") {
  throw new Error("dashboard TypeScript plugin contract failed");
}

const install = pluginInstallRequest("dev", "auth", "1.0.0");
const grant = pluginGrantRequest("dev", "auth", "1.0.0", "auth.manage");
const enable = pluginEnableRequest("dev", "auth", "1.0.0");
if (install.params.operation !== "plugin.install" || grant.params.operation !== "plugin.grant" || enable.params.operation !== "plugin.enable") {
  throw new Error("dashboard lifecycle request contract failed");
}

const result = handle(projectCreateRequest("dev"));
if (typeof result !== "object") throw new Error("dashboard handler did not return a result");
