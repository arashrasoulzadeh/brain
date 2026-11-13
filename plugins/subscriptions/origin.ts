export function validateBrowserOrigin(origin: string, allowedOrigins: readonly string[]): void {
  if (!origin || !allowedOrigins.includes(origin)) throw new Error("origin denied");
  let parsed: URL;
  try { parsed = new URL(origin); } catch { throw new Error("origin denied"); }
  if (!["http:", "https:"].includes(parsed.protocol) || parsed.username || parsed.password || parsed.pathname !== "/" || parsed.search || parsed.hash) throw new Error("origin denied");
}
