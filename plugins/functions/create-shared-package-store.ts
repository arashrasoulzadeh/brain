import { createProjectMap, validateProjectId } from "../sdk-typescript/src/index.js";

export function createSharedPackageStore() {
  const packages = createProjectMap<string>();
  return {
    save(projectId: string, packageId: string, source: string): void {
      validateProjectId(projectId);
      if (!packageId || !source) throw new Error("invalid shared package");
      packages.set(projectId, packageId, source);
    },
    load(projectId: string, packageId: string): string {
      const source = packages.get(projectId, packageId);
      if (!source) throw new Error("shared package not found");
      return source;
    },
  };
}
