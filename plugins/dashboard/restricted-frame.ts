export function restrictedFrame(origin: string, projectId: string, viewPath: string) {
  const parsed = new URL(origin);
  if (parsed.protocol !== "https:" || !projectId.trim() || !viewPath.startsWith("/"))
    throw new Error("invalid restricted frame");
  return { src: `${parsed.origin}${viewPath}`, projectId, sandbox: "allow-scripts" } as const;
}
