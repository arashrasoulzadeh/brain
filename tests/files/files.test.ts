import { fileRequest, handle, manifest } from "../../plugins/files/src/index.js";

const request = fileRequest("dev", "files.upload", "avatar", { contentType: "image/png" });
if (manifest.id !== "files" || request.params.operation !== "files.upload") throw new Error("files plugin contract failed");
if (typeof handle(request) !== "object") throw new Error("files handler contract failed");
