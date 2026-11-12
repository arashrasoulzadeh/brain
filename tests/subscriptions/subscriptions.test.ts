import { createHandler, subscriptionRequest } from "../../plugins/subscriptions/index.js";

const request = subscriptionRequest("dev", "subscriptions", "0.1.0", "subscriptions.open", "records.users");
if (request.params.pluginId !== "subscriptions" || request.params.operation !== "subscriptions.open") throw new Error("subscriptions plugin contract failed");
if (typeof createHandler([{ operation: "subscriptions.open", function: "open", capability: "subscriptions.open" }])(request) !== "object") throw new Error("subscriptions handler contract failed");
