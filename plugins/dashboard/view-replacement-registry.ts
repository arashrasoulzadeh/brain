export function createViewReplacementRegistry() {
  const replacements = new Map<string, string>();
  return {
    activate(projectId: string, view: string, pluginId: string) {
      if (!projectId.trim() || !view.trim() || !pluginId.trim()) throw new Error("invalid view replacement");
      const key = `${projectId}:${view}`;
      if (replacements.has(key)) throw new Error("view replacement already active");
      replacements.set(key, pluginId);
      return { projectId, view, pluginId };
    },
    active(projectId: string, view: string) { return replacements.get(`${projectId}:${view}`); },
  };
}
