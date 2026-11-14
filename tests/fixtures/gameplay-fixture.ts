export type GameplayFixture = {
  projectId: string;
  roomId: string;
  players: readonly string[];
  initialRevision: 0;
  firstMove: { playerId: string; cell: 0 };
};

export function createGameplayFixture(projectId = "dev"): GameplayFixture {
  if (!projectId.trim()) throw new Error("fixture project required");
  return {
    projectId,
    roomId: "fixture-room",
    players: ["fixture-player-a", "fixture-player-b"],
    initialRevision: 0,
    firstMove: { playerId: "fixture-player-a", cell: 0 },
  };
}
