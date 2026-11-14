import type { RpcRequest } from "./index.js";
import { request } from "./request.js";
export function pluginRequest(prefix: string, projectId: string, pluginId: string, version: string, operation: string, payload: unknown = {}, identity = ""): RpcRequest { return request(`${prefix}-${operation}-${projectId}${identity ? `-${identity}` : ""}`, { protocol: 1, projectId, pluginId, version, operation, payload }); }
