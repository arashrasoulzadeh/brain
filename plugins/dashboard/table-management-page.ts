export function tableManagementPage(
  projectId: string,
  tables: readonly string[],
): { projectId: string; title: string; tables: readonly string[] } {
  if (!projectId || tables.some((table) => !/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(table)))
    throw new Error("invalid table management");
  return { projectId, title: "Table management", tables: [...tables] };
}
