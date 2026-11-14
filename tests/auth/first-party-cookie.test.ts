import { firstPartySessionCookie } from "../../plugins/auth/index.js";

const cookie = firstPartySessionCookie("session-token", 3600);
if (
  cookie !==
  "__Host-brain_session=session-token; Max-Age=3600; Path=/; HttpOnly; Secure; SameSite=Lax"
)
  throw new Error("first-party session cookie is not secure");

try {
  firstPartySessionCookie("", 3600);
  throw new Error("empty session cookie accepted");
} catch (error) {
  if (!(error instanceof Error) || error.message !== "invalid session cookie") throw error;
}
