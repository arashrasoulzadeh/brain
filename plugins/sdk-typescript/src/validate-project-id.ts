export function validateProjectId(projectId: string, errorMessage = "invalid project id"): string {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(projectId)) throw new Error(errorMessage);
  return projectId;
}
