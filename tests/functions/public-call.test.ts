import { call } from "../../plugins/functions/index.js";

const accepted = call({
  name: "hello",
  version: "2",
  input: { ok: true },
  permission: "public",
}) as Record<string, unknown>;
if (accepted.route !== "public" || accepted.version !== "2" || accepted.name !== "hello")
  throw new Error("public function call route failed");

try {
  call({ name: "private", permission: "functions.invoke" });
  throw new Error("private function was exposed through public route");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "function is not public") throw error;
}
