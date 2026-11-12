import { createInterface } from "node:readline";
import type { RpcRequest, RpcResponse } from "./index.js";

export type PluginHandler = (request: RpcRequest) => Promise<unknown> | unknown;

export function startWorker(handler: PluginHandler): void {
  const input = createInterface({ input: process.stdin, crlfDelay: Infinity });
  input.on("line", async (line) => {
    if (Buffer.byteLength(line, "utf8") > 1024 * 1024) {
      process.stdout.write(JSON.stringify({ jsonrpc: "2.0", id: "", errorCode: "payload_too_large" }) + "\n");
      return;
    }
    let request: RpcRequest;
    try { request = JSON.parse(line) as RpcRequest; }
    catch { process.stdout.write(JSON.stringify({ jsonrpc: "2.0", id: "", errorCode: "invalid_json" }) + "\n"); return; }
    try {
      const result = await handler(request);
      const response: RpcResponse = { jsonrpc: "2.0", id: request.id, result: { protocol: request.params.protocol, success: true, payload: result } };
      process.stdout.write(JSON.stringify(response) + "\n");
    } catch {
      process.stdout.write(JSON.stringify({ jsonrpc: "2.0", id: request.id, errorCode: "plugin_error" }) + "\n");
    }
  });
}
