type ProjectRequestKind = "catalog" | "functions" | "playground";

/** Only the newest request for the currently selected project may publish data. */
export function createProjectRequestGate() {
  let projectID = "";
  let epoch = 0;
  const sequences: Record<ProjectRequestKind, number> = { catalog: 0, functions: 0, playground: 0 };
  return {
    current: () => projectID,
    select(next: string) {
      if (projectID === next) return false;
      projectID = next;
      epoch++;
      return true;
    },
    begin(project: string, kind: ProjectRequestKind): (() => boolean) | undefined {
      if (project !== projectID) return undefined;
      const started = epoch;
      const sequence = ++sequences[kind];
      return () => started === epoch && sequence === sequences[kind];
    },
  };
}
