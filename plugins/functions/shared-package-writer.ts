import { validateProjectId } from "../sdk-typescript/src/index.js";
import type { SharedPackageState } from "./shared-package-state.js";
export class SharedPackageWriter {
  constructor(private readonly state: SharedPackageState) {}
  save(projectId: string, packageId: string, source: string): void {
    validateProjectId(projectId);
    if (!packageId || !source) throw new Error("invalid shared package");
    this.state.packages.set(`${projectId}:${packageId}`, source);
  }
}
