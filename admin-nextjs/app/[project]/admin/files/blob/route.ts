import type { NextRequest } from "next/server";

// Explicit streaming proxy: unlike generic rewrites, never clone the upload
// body or materialize an entire download. The Go host performs authorization.
export const runtime = "nodejs";
export const dynamic = "force-dynamic";
type Context = { params: Promise<{ project: string }> };
async function transfer(request: NextRequest, context: Context): Promise<Response> {
  const { project } = await context.params;
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]*$/.test(project) || project === "_system") return new Response("Project unavailable", { status: 404 });
  const origin = process.env.BRAIN_API_ORIGIN ?? "http://127.0.0.1:18082";
  const url = new URL("/" + encodeURIComponent(project) + "/admin/files/blob", origin);
  url.search = request.nextUrl.search;
  const headers = new Headers();
  for (const name of ["authorization", "content-type", "content-length", "x-project-id", "x-project-slug"]) {
    const value = request.headers.get(name);
    if (value !== null) headers.set(name, value);
  }
  try {
    const init: RequestInit & { duplex?: "half" } = { method: request.method, headers, cache: "no-store", redirect: "error", signal: request.signal };
    if (request.method === "POST") { init.body = request.body; init.duplex = "half"; }
    else { headers.delete("content-length"); }
    const response = await fetch(url, init);
    const outgoing = new Headers({ "Cache-Control": "no-store", "X-Content-Type-Options": "nosniff" });
    for (const name of ["content-type", "content-disposition", "content-length"]) {
      const value = response.headers.get(name);
      if (value !== null) outgoing.set(name, value);
    }
    return new Response(response.body, { status: response.status, headers: outgoing });
  } catch {
    return Response.json({ status: "error", message: "File transfer unavailable" }, { status: 502 });
  }
}
export const POST = transfer;
export const GET = transfer;
