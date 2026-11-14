export function createViewReplacementRegistry() {
  const replacements = new Map<string, string>();
  return {
    activate(projectId: string, view: string, pluginId: string) {
      if (!projectId.trim() || !view.trim() || !pluginId.trim())
        throw new Error("invalid view replacement");
      const key = `${projectId}:${view}`;
      if (replacements.has(key)) throw new Error("view replacement already active");
      replacements.set(key, pluginId);
      return { projectId, view, pluginId };
    },
    active(projectId: string, view: string) {
      return replacements.get(`${projectId}:${view}`);
    },
    submit(
      hostProjectId: string,
      projectId: string,
      view: string,
      authorize: (projectId: string, view: string) => boolean,
      audit: (projectId: string, view: string, pluginId: string) => void,
    ) {
      if (!hostProjectId.trim() || hostProjectId !== projectId)
        throw new Error("host project context mismatch");
      if (!authorize(projectId, view)) throw new Error("admin view authorization denied");
      const pluginId = replacements.get(`${projectId}:${view}`);
      if (!pluginId) throw new Error("view replacement not found");
      audit(projectId, view, pluginId);
      return { projectId, view, pluginId };
    },
  };
}
