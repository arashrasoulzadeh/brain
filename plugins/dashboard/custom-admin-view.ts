export function customAdminView(projectId: string) {
  if (!projectId.trim()) throw new Error("project identity required");
  return { contentType: "text/html", projectId, html: `<main data-project="${projectId}"><h1>Custom admin view</h1></main>` } as const;
}
