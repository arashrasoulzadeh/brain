import type { SharedPackageState } from "./shared-package-state.js";
import { validateProjectId } from "../sdk-typescript/src/index.js";
export class SharedPackageReader {
  constructor(private readonly state: SharedPackageState) {}
  load(projectId: string, packageId: string): string {
    validateProjectId(projectId);
    const source = this.state.packages.get(`${projectId}:${packageId}`);
    if (!source) throw new Error("shared package not found");
    return source;
  }
}
