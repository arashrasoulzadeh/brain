import { createGameplayFixture } from "./gameplay-fixture.js";

const fixture = createGameplayFixture();
if (
  fixture.projectId !== "dev" ||
  fixture.players.length !== 2 ||
  fixture.initialRevision !== 0 ||
  fixture.firstMove.cell !== 0
)
  throw new Error("shared gameplay fixture is incomplete");
