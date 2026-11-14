import { createSubscriptionFixture } from "./subscription-fixture.js";

const fixture = createSubscriptionFixture();
if (
  fixture.projectId !== "dev" ||
  fixture.channel !== "records.profiles" ||
  fixture.event.operation !== "updated"
)
  throw new Error("shared subscription fixture is incomplete");
