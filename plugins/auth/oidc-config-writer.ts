import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { OIDCConfig } from "./index.js";

export class OIDCConfigWriter {
  constructor(private readonly configs: Map<string, OIDCConfig>) {}
  set(projectId: string, config: OIDCConfig): void {
    validateProjectId(projectId);
    if (
      !config.audience ||
      !config.clientId ||
      !/^https:\/\//.test(config.issuer) ||
      !/^https:\/\//.test(config.redirectUri)
    )
      throw new Error("invalid OIDC configuration");
    this.configs.set(projectId, { ...config });
  }
}
