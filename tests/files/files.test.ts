import { AuthorizedUploadStream, configurePrivateBucket, FileAccessPolicy, FileQuotaLedger, fileRequest, PendingUploadStore, projectObjectKey } from "../../plugins/files/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = fileRequest("dev", "files", "0.1.0", "files.upload", "avatar", { contentType: "image/png" });
if (request.params.pluginId !== "files" || request.params.operation !== "files.upload") throw new Error("files plugin contract failed");
if (typeof createRouteHandler([{ operation: "files.upload", function: "upload", capability: "files.upload" }])(request) !== "object") throw new Error("files handler contract failed");
if (configurePrivateBucket({ endpoint: "https://s3.example.com", bucket: "brain", region: "us-east-1" }).publicAccess) throw new Error("private bucket exposed");
try { configurePrivateBucket({ endpoint: "http://s3.example.com", bucket: "brain", region: "us-east-1" }); throw new Error("insecure bucket accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "private bucket requires HTTPS") throw error; }
if (projectObjectKey("dev", "avatar.png") !== "projects/dev/objects/avatar.png" || projectObjectKey("prod", "avatar.png") === projectObjectKey("dev", "avatar.png")) throw new Error("project object namespace failed");
try { projectObjectKey("dev/other", "avatar.png"); throw new Error("cross-project object key accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid object identity") throw error; }
const uploads = new PendingUploadStore();
if (uploads.create("dev", "avatar.png", "image/png", 12).status !== "pending" || uploads.get("prod", "avatar.png")) throw new Error("pending upload metadata isolation failed");
if (uploads.publish("dev", "avatar.png", "etag-1").status !== "published" || uploads.get("dev", "avatar.png")?.etag !== "etag-1") throw new Error("published file metadata failed");
try { uploads.publish("dev", "missing.png", "etag-2"); throw new Error("missing object published"); } catch (error) { if (!(error instanceof Error) || error.message !== "object upload not confirmed") throw error; }
const quota = new FileQuotaLedger(new Map([["dev", 20]]));
quota.reserve("dev", 12);
try { quota.reserve("dev", 9); throw new Error("quota exceeded upload accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "file quota exceeded") throw error; }
const stream = new AuthorizedUploadStream();
stream.authorize("dev", "avatar.png", "user-1");
if (stream.write("dev", "avatar.png", "user-1", new Uint8Array(3)) !== 3) throw new Error("authorized upload failed");
try { stream.write("dev", "avatar.png", "user-2", new Uint8Array(1)); throw new Error("unauthorized upload accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "upload denied") throw error; }
const access = new FileAccessPolicy();
access.grant("dev", "avatar.png", "user-1"); access.read("dev", "avatar.png", "user-1"); access.delete("dev", "avatar.png", "user-1");
try { access.read("dev", "avatar.png", "user-2"); throw new Error("unauthorized file read accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "file access denied") throw error; }
