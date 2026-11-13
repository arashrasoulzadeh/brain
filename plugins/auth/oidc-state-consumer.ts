import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { OIDCState } from "./oidc-state-issuer.js";

export class OIDCStateConsumer {
  constructor(private readonly states: Map<string, OIDCState>) {}
  consume(state: string, projectId: string, clientId: string, redirectUri: string): void { validateProjectId(projectId); const record = this.states.get(state); if (!record || record.projectId !== projectId || record.clientId !== clientId || record.redirectUri !== redirectUri) throw new Error("callback state mismatch"); this.states.delete(state); }
}
