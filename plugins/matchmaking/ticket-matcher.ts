import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { QueueRuntime } from "./queue-runtime.js";
export class TicketMatcher { constructor(private readonly runtime: QueueRuntime) {} match(projectId: string, scope: string, matchSize: number) { validateProjectId(projectId); const key = `${projectId}:${scope}`; const queue = this.runtime.queues.get(key) ?? []; if (matchSize < 1 || queue.length < matchSize) return []; const matched = queue.slice(0, matchSize); this.runtime.queues.set(key, queue.slice(matchSize)); return matched.map((ticket) => ({ ...ticket })); } }
