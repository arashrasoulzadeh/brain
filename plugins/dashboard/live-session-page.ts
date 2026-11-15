import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { LiveSessionSummary } from "./types.js";

export function liveSessionPage(projectId: string, sessions: readonly LiveSessionSummary[]) {
  validateProjectId(projectId, "invalid live sessions");
  if (
    sessions.some(
      (item) =>
        !item.id ||
        !item.kind ||
        !item.status ||
        !Number.isSafeInteger(item.participantCount) ||
        item.participantCount < 0,
    )
  )
    throw new Error("invalid live sessions");
  return { projectId, title: "Live sessions", sessions: sessions.map((item) => ({ ...item })) };
}
