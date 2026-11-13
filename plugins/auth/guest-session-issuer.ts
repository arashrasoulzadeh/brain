import { randomBytes } from "node:crypto";
import { validateProjectId } from "../sdk-typescript/src/index.js";
import { newSession, projectUserID, type AuthState } from "./auth-state.js";
export class GuestSessionIssuer { constructor(private readonly state: AuthState) {} guest(projectId: string) { validateProjectId(projectId); const username = `guest-${randomBytes(8).toString("hex")}`; const session = newSession(this.state, projectId, username, Date.now()); return { ...session, userId: projectUserID(projectId, username), guest: true }; } }
