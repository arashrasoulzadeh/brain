import { handle, manifest, subscriptionRequest } from "../../plugins/subscriptions/src/index.js";

const request = subscriptionRequest("dev", "subscriptions.open", "records.users");
if (manifest.id !== "subscriptions" || request.params.operation !== "subscriptions.open") throw new Error("subscriptions plugin contract failed");
if (typeof handle(request) !== "object") throw new Error("subscriptions handler contract failed");
