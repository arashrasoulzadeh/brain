import { configurePrivateBucket, fileRequest } from "../../plugins/files/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = fileRequest("dev", "files", "0.1.0", "files.upload", "avatar", { contentType: "image/png" });
if (request.params.pluginId !== "files" || request.params.operation !== "files.upload") throw new Error("files plugin contract failed");
if (typeof createRouteHandler([{ operation: "files.upload", function: "upload", capability: "files.upload" }])(request) !== "object") throw new Error("files handler contract failed");
if (configurePrivateBucket({ endpoint: "https://s3.example.com", bucket: "brain", region: "us-east-1" }).publicAccess) throw new Error("private bucket exposed");
try { configurePrivateBucket({ endpoint: "http://s3.example.com", bucket: "brain", region: "us-east-1" }); throw new Error("insecure bucket accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "private bucket requires HTTPS") throw error; }
