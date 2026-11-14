import type { FunctionDraft, FunctionPublication } from "./index.js";
export function publishSharedHelperArtifact(
  projectId: string,
  packageId: string,
  source: string,
  artifactDigest: string,
  toolchain: string,
  publishedAt: number,
  publish: (
    draft: FunctionDraft,
    digest: string,
    compiler: string,
    at: number,
  ) => FunctionPublication,
): FunctionPublication {
  if (!packageId || !source.includes("package helper"))
    throw new Error("invalid shared helper artifact");
  return publish(
    {
      projectId,
      functionId: packageId,
      source: `package main\n\n// helper: ${source}`,
      savedAt: publishedAt,
    },
    artifactDigest,
    toolchain,
    publishedAt,
  );
}
