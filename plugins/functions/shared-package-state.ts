export type SharedPackageState = { packages: Map<string, string> };
export function createSharedPackageState(): SharedPackageState { return { packages: new Map() }; }
