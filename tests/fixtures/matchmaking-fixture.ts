export type MatchmakingFixture = {
  projectId: string;
  queue: "ranked";
  players: readonly [{ id: string; rating: number }, { id: string; rating: number }];
};

export function createMatchmakingFixture(projectId = "dev"): MatchmakingFixture {
  if (!projectId.trim()) throw new Error("fixture project required");
  return {
    projectId,
    queue: "ranked",
    players: [
      { id: "fixture-player-a", rating: 1000 },
      { id: "fixture-player-b", rating: 1000 },
    ],
  };
}
