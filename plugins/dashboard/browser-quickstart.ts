export function browserQuickstart(
  projectId: string,
  apiOrigin: string,
): { projectId: string; apiOrigin: string; steps: readonly string[] } {
  const origin = new URL(apiOrigin);
  if (!projectId || !["http:", "https:"].includes(origin.protocol))
    throw new Error("invalid browser quickstart");
  return {
    projectId,
    apiOrigin: origin.origin,
    steps: ["create-session", "load-project", "open-dashboard"],
  };
}
