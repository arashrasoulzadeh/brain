import { randomBytes } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";

export type OIDCState = { projectId: string; clientId: string; redirectUri: string };
export class OIDCStateIssuer {
  constructor(private readonly states: Map<string, OIDCState>) {}
  issue(projectId: string, clientId: string, redirectUri: string): string {
    validateProjectId(projectId);
    const state = randomBytes(24).toString("base64url");
    this.states.set(state, { projectId, clientId, redirectUri });
    return state;
  }
}
