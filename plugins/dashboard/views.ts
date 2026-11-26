import { validateProjectId } from "../sdk-typescript/src/index.js";

export function hostOwnedAdminView(projectId: string) {
  validateProjectId(projectId, "project identity required");
  return {
    kind: "host-admin-view",
    projectId,
    title: "Project administration",
    trusted: true,
  } as const;
}

export function projectOverviewContribution(projectId: string) {
  validateProjectId(projectId, "project identity required");
  return {
    view: "project.overview",
    projectId,
    title: "Project overview",
    replace: false,
  } as const;
}

export function pluginHealthContribution(projectId: string) {
  validateProjectId(projectId, "project identity required");
  return { view: "plugin.health", projectId, title: "Plugin health", replace: false } as const;
}

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

export function customAdminView(projectId: string) {
  validateProjectId(projectId);
  return {
    contentType: "text/html",
    projectId,
    html: `<main data-project="${projectId}"><h1>Custom admin view</h1></main>`,
  } as const;
}
