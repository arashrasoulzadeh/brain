import { validateProjectId } from "./validate-project-id.js";

// In-memory partitioning only; callers still require host-authenticated project context.
export function createProjectMap<Value>() {
  const projects = new Map<string, Map<string, Value>>();

  return {
    get(projectId: string, key: string): Value | undefined {
      return projects.get(validateProjectId(projectId))?.get(key);
    },
    set(projectId: string, key: string, value: Value): void {
      validateProjectId(projectId);
      let entries = projects.get(projectId);
      if (!entries) {
        entries = new Map();
        projects.set(projectId, entries);
      }
      entries.set(key, value);
    },
  };
}
