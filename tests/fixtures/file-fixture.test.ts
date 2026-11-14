import { createFileFixture } from "./file-fixture.js";

const fixture = createFileFixture();
if (
  fixture.projectId !== "dev" ||
  !fixture.bucket ||
  !fixture.objectKey.includes("/") ||
  fixture.contentType !== "text/plain"
)
  throw new Error("shared file fixture is incomplete");
