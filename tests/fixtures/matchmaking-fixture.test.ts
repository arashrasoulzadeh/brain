import { createMatchmakingFixture } from "./matchmaking-fixture.js";

const fixture = createMatchmakingFixture();
if (
  fixture.projectId !== "dev" ||
  fixture.queue !== "ranked" ||
  fixture.players.length !== 2 ||
  fixture.players[0].id === fixture.players[1].id
)
  throw new Error("shared matchmaking fixture is incomplete");
