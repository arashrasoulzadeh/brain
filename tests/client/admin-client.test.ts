import { adminAction, adminLogin } from "../../clients/javascript/src/index.js";

const calls: RequestInit[] = [];
const urls: string[] = [];
const request = async (_url: string | URL | Request, init?: RequestInit) => {
  calls.push(init ?? {});
  urls.push(String(_url));
  return new Response(JSON.stringify({ session: "operator-session" }), { status: 200 });
};
const login = await adminLogin("http://localhost:18081", "admin", "secret", request);
if (login.session !== "operator-session") throw new Error("admin login client failed");
await adminAction(
  "http://localhost:18081",
  login.session,
  { operation: "plugin.enable", project_id: "dev" },
  request,
);
if (
  calls.length !== 2 ||
  new Headers(calls[1]?.headers).get("authorization") !== "Bearer operator-session"
)
  throw new Error("admin action client failed");

if (urls[0] !== "http://localhost:18081/_system/admin/session" ||
    urls[1] !== "http://localhost:18081/dev/admin/actions") {
  throw new Error("operator URL must be project or system scoped");
}
