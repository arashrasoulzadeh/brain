export type ReconnectFixture = {
  projectId: string;
  sessionId: string;
  lastEventSequence: 4;
  reconnectToken: string;
};

export function createReconnectFixture(projectId = "dev"): ReconnectFixture {
  if (!projectId.trim()) throw new Error("fixture project required");
  return {
    projectId,
    sessionId: "fixture-session",
    lastEventSequence: 4,
    reconnectToken: "fixture-reconnect-token",
  };
}
