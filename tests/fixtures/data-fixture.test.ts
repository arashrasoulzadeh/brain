import { createDataFixture } from "./data-fixture.js";

const fixture = createDataFixture();
if (
  fixture.projectId !== "dev" ||
  fixture.collection !== "profiles" ||
  fixture.record.id !== "fixture-user"
)
  throw new Error("shared data fixture is incomplete");
