import { SubscriptionTicketStore, subscriptionRequest, validateBrowserOrigin } from "../../plugins/subscriptions/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = subscriptionRequest("dev", "subscriptions", "0.1.0", "subscriptions.open", "records.users");
if (request.params.pluginId !== "subscriptions" || request.params.operation !== "subscriptions.open") throw new Error("subscriptions plugin contract failed");
if (typeof createRouteHandler([{ operation: "subscriptions.open", function: "open", capability: "subscriptions.open" }])(request) !== "object") throw new Error("subscriptions handler contract failed");
const tickets = new SubscriptionTicketStore();
const ticket = tickets.issue("dev", "records.users", "subscriptions.open", 1000, 30_000);
if (tickets.validate(ticket.ticket, "dev", "records.users", "subscriptions.open", 2000).projectId !== "dev") throw new Error("subscription ticket validation failed");
try { tickets.validate(ticket.ticket, "prod", "records.users", "subscriptions.open", 2000); throw new Error("cross-project ticket accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid subscription ticket") throw error; }
const oneUse = tickets.issue("dev", "records.orders", "subscriptions.open", 1000);
tickets.consume(oneUse.ticket, "dev", "records.orders", "subscriptions.open", 2000);
try { tickets.consume(oneUse.ticket, "dev", "records.orders", "subscriptions.open", 2001); throw new Error("replayed subscription ticket accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid subscription ticket") throw error; }
validateBrowserOrigin("https://app.example.com/", ["https://app.example.com/"]);
try { validateBrowserOrigin("https://evil.example.com/", ["https://app.example.com/"]); throw new Error("browser origin accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "origin denied") throw error; }
