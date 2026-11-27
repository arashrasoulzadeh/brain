import { createResourceRequest } from "../sdk-typescript/src/index.js";

export const subscriptionRequest = createResourceRequest<"subscriptions.open" | "subscriptions.receive" | "subscriptions.resync" | "subscriptions.close">("subscriptions", "channel");
