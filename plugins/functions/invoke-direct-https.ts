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
  const normalizedIPv6Host = hostname.replace(/^\[|\]$/g, "");
  const addressType = isIP(normalizedIPv6Host);
  const privateIPv4 =
    addressType === 4 &&
    (/^(10|127)\./.test(normalizedIPv6Host) ||
      /^192\.168\./.test(normalizedIPv6Host) ||
      /^172\.(1[6-9]|2[0-9]|3[0-1])\./.test(normalizedIPv6Host) ||
      normalizedIPv6Host === "169.254.169.254");
  const privateIPv6 =
    addressType === 6 &&
    (normalizedIPv6Host === "::1" ||
      normalizedIPv6Host.startsWith("::ffff:") ||
      normalizedIPv6Host.startsWith("fc") ||
      normalizedIPv6Host.startsWith("fd") ||
      normalizedIPv6Host.startsWith("fe8") ||
      normalizedIPv6Host.startsWith("fe9") ||
      normalizedIPv6Host.startsWith("fea") ||
      normalizedIPv6Host.startsWith("feb"));
  const ambiguousNumericHost = addressType === 0 && /^[0-9.]+$/.test(hostname);
  if (!hostname || hostname === "localhost" || privateIPv4 || privateIPv6 || ambiguousNumericHost)
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
