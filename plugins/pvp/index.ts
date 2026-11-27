import { requireProjectInput, type PluginContext } from "../sdk-typescript/src/index.js";

export { CommandDeduplicator, assertMatchMember, nextTurn, createRevisionRuntime, commitRevision, replayHistory, expireTurn, pauseTurn, resumeTurn, privatePlayerView, parseMatchState } from "../sdk-typescript/src/index.js";
export type { TurnMatchState, RevisionRuntime, CommittedCommand } from "../sdk-typescript/src/index.js";

export function describe(context: PluginContext, input: unknown) {
  requireProjectInput(context, input);
  return { projectId: context.project.id, turnBased: true, revisionedCommands: true };
}
