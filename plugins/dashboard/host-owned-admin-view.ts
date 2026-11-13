export function hostOwnedAdminView(projectId: string) {
  if (!projectId.trim()) throw new Error("project identity required");
  return { kind: "host-admin-view", projectId, title: "Project administration", trusted: true } as const;
}
