import type { MatchmakingConfigReader } from "./matchmaking-config-reader.js";
import type { MatchmakingConfigStatus } from "./matchmaking-config-status.js";
import { createQueueRuntime } from "./queue-runtime.js";
import { TicketEnqueuer } from "./ticket-enqueuer.js";
import { TicketReader } from "./ticket-reader.js";
import { TicketMatcher } from "./ticket-matcher.js";
import { TicketExpirer } from "./ticket-expirer.js";
import { TicketRequeuer } from "./ticket-requeuer.js";
import { AssignmentCreator } from "./assignment-creator.js";
import { TicketCanceller } from "./ticket-canceller.js";
import { ScopeTicketCanceller } from "./scope-ticket-canceller.js";
export function createMatchmakingQueue(configs?: { reader: MatchmakingConfigReader; status: MatchmakingConfigStatus }) { const runtime = createQueueRuntime(configs); const enqueue = new TicketEnqueuer(runtime); const reader = new TicketReader(runtime); const matcher = new TicketMatcher(runtime); const expirer = new TicketExpirer(runtime); const requeuer = new TicketRequeuer(runtime); const assigner = new AssignmentCreator(runtime); const canceller = new TicketCanceller(runtime); const scopeCanceller = new ScopeTicketCanceller(runtime); return { enqueue: enqueue.enqueue.bind(enqueue), peek: reader.peek.bind(reader), match: matcher.match.bind(matcher), expire: expirer.expire.bind(expirer), requeue: requeuer.requeue.bind(requeuer), assign: assigner.assign.bind(assigner), cancelIncompatible: canceller.cancelIncompatible.bind(canceller), cancelScope: scopeCanceller.cancelScope.bind(scopeCanceller), setLifecycleRevision: (revision: number) => { runtime.lifecycleRevision = revision; } }; }
