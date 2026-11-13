export function fileManagementPage(projectId: string, files: readonly { objectId: string; size: number; status: string }[]): { projectId: string; title: string; files: readonly { objectId: string; size: number; status: string }[] } {
  if (!projectId || files.some((file) => !file.objectId || !Number.isSafeInteger(file.size) || file.size < 0)) throw new Error("invalid file management");
  return { projectId, title: "File management", files: files.map((file) => ({ ...file })) };
}
