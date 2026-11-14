import { isIP } from "node:net";
import type { FunctionInvocationResult, HttpsFetcher } from "./invocation-types.js";
import { parseInvocationResult } from "./parse-invocation-result.js";
export async function invokeDirectHttps(
  endpoint: string,
  input: unknown,
  fetcher: HttpsFetcher = fetch,
): Promise<FunctionInvocationResult> {
  const target = new URL(endpoint);
  if (target.protocol !== "https:" || target.username || target.password)
    throw new Error("HTTPS endpoint required");
  const hostname = target.hostname.toLowerCase();
  const addressType = isIP(hostname);
  const privateIPv4 =
    addressType === 4 &&
    (/^(10|127)\./.test(hostname) ||
      /^192\.168\./.test(hostname) ||
      /^172\.(1[6-9]|2[0-9]|3[0-1])\./.test(hostname) ||
      hostname === "169.254.169.254");
  const privateIPv6 =
    addressType === 6 &&
    (hostname === "::1" ||
      hostname.startsWith("fc") ||
      hostname.startsWith("fd") ||
      hostname.startsWith("fe8") ||
      hostname.startsWith("fe9") ||
      hostname.startsWith("fea") ||
      hostname.startsWith("feb"));
  if (!hostname || hostname === "localhost" || privateIPv4 || privateIPv6)
    throw new Error("HTTPS endpoint target denied");
  const response = await fetcher(target.toString(), {
    method: "POST",
    redirect: "error",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(input),
  });
  if (!response.ok) throw new Error(`HTTPS invocation failed: ${response.status}`);
  return parseInvocationResult(await response.json());
}
