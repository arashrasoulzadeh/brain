export type WorkflowNode = { id: string; next: readonly string[] };
export function validateWorkflowGraph(nodes: readonly WorkflowNode[]): void {
  const byId = new Map(nodes.map((node) => [node.id, node]));
  if (
    byId.size !== nodes.length ||
    nodes.some((node) => !node.id || node.next.some((next) => !byId.has(next)))
  )
    throw new Error("invalid workflow graph");
  const visiting = new Set<string>();
  const visited = new Set<string>();
  const visit = (id: string): void => {
    if (visiting.has(id)) throw new Error("workflow cycle");
    if (visited.has(id)) return;
    visiting.add(id);
    for (const next of byId.get(id)!.next) visit(next);
    visiting.delete(id);
    visited.add(id);
  };
  for (const node of nodes) visit(node.id);
}
