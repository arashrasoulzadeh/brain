import { InitialFetchBuffer, SlowConsumerQueue, SubscriptionGrantStore, SubscriptionAuthorization, SubscriptionOwnerInvalidator, SubscriptionEventAppender, SubscriptionEventReader, SubscriptionTicketIssuer, SubscriptionTicketValidator, SubscriptionTicketConsumer, subscriptionRequest, validateBrowserOrigin } from "../../plugins/subscriptions/index.js";
import { createRouteHandler } from "../../plugins/sdk-typescript/src/index.js";

const request = subscriptionRequest("dev", "subscriptions", "0.1.0", "subscriptions.open", "records.users");
if (request.params.pluginId !== "subscriptions" || request.params.operation !== "subscriptions.open") throw new Error("subscriptions plugin contract failed");
if (typeof createRouteHandler([{ operation: "subscriptions.open", function: "open", capability: "subscriptions.open" }])(request) !== "object") throw new Error("subscriptions handler contract failed");
const ticketMap = new Map(); const ticketIssuer = new SubscriptionTicketIssuer(ticketMap); const ticketValidator = new SubscriptionTicketValidator(ticketMap); const ticketConsumer = new SubscriptionTicketConsumer(ticketMap);
const ticket = ticketIssuer.issue("dev", "records.users", "subscriptions.open", 1000, 30_000);
if (ticketValidator.validate(ticket.ticket, "dev", "records.users", "subscriptions.open", 2000).projectId !== "dev") throw new Error("subscription ticket validation failed");
try { ticketValidator.validate(ticket.ticket, "prod", "records.users", "subscriptions.open", 2000); throw new Error("cross-project ticket accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid subscription ticket") throw error; }
const oneUse = ticketIssuer.issue("dev", "records.orders", "subscriptions.open", 1000);
ticketConsumer.consume(oneUse.ticket, "dev", "records.orders", "subscriptions.open", 2000);
try { ticketConsumer.consume(oneUse.ticket, "dev", "records.orders", "subscriptions.open", 2001); throw new Error("replayed subscription ticket accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid subscription ticket") throw error; }
validateBrowserOrigin("https://app.example.com/", ["https://app.example.com/"]);
try { validateBrowserOrigin("https://evil.example.com/", ["https://app.example.com/"]); throw new Error("browser origin accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "origin denied") throw error; }
const grants = new Set<string>();
try { new SubscriptionGrantStore(grants).grant("dev:other", "records.users", "user-1"); throw new Error("malformed project grant accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "invalid subscription grant") throw error; }
new SubscriptionGrantStore(grants).grant("dev", "records.users", "user-1");
new SubscriptionAuthorization(grants).authorize("dev", "records.users", "user-1");
try { new SubscriptionAuthorization(grants).authorize("dev", "records.users", "user-2"); throw new Error("unauthorized subscription accepted"); } catch (error) { if (!(error instanceof Error) || error.message !== "subscription denied") throw error; }
new SubscriptionOwnerInvalidator(grants).invalidate("dev", "user-1");
try { new SubscriptionAuthorization(grants).authorize("dev", "records.users", "user-1"); throw new Error("old owner subscription survived transfer"); } catch (error) { if (!(error instanceof Error) || error.message !== "subscription denied") throw error; }
const buffer = new InitialFetchBuffer<string>();
buffer.push("write-during-fetch");
if (buffer.complete(["snapshot"]).join(",") !== "snapshot,write-during-fetch") throw new Error("initial fetch write lost");
const queue = new SlowConsumerQueue<string>(1);
if (!queue.push("event-1") || queue.push("event-2") || queue.size !== 1 || queue.shift() !== "event-1") throw new Error("slow consumer queue failed");
const replayEvents: { sequence: number; value: string }[] = []; let replaySequence = 0;
const replayAppender = new SubscriptionEventAppender(replayEvents, () => ++replaySequence);
replayAppender.append("one"); replayAppender.append("two");
if (new SubscriptionEventReader(replayEvents).since(1).map((event) => event.value).join(",") !== "two") throw new Error("subscription replay failed");
