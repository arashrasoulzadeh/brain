import type { OIDCConfig } from "./index.js";

export class OIDCConfigReader {
  constructor(private readonly configs: ReadonlyMap<string, OIDCConfig>) {}
  get(projectId: string): OIDCConfig { const config = this.configs.get(projectId); if (!config) throw new Error("OIDC configuration not found"); return { ...config }; }
}
