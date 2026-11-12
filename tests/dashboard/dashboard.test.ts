import { manifest, projectCreateRequest } from "../../plugins/dashboard/src/index.js";

const request = projectCreateRequest("dev");
if (manifest.id !== "dashboard" || request.params.operation !== "project.create") {
  throw new Error("dashboard TypeScript plugin contract failed");
}
