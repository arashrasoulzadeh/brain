import { createReconnectFixture } from "./reconnect-fixture.js";

const fixture = createReconnectFixture();
if (
  fixture.projectId !== "dev" ||
  !fixture.sessionId ||
  fixture.lastEventSequence !== 4 ||
  !fixture.reconnectToken
)
  throw new Error("shared reconnect fixture is incomplete");
