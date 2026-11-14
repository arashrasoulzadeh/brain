import { createFunctionFixture } from "./function-fixture.js";

const fixture = createFunctionFixture();
if (
  fixture.projectId !== "dev" ||
  fixture.functionId !== "hello" ||
  !fixture.version ||
  !fixture.inputs.name ||
  !fixture.idempotencyKey
)
  throw new Error("shared function fixture is incomplete");
