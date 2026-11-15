import { validateProjectId } from "../sdk-typescript/src/index.js";

export function fileManagementPage(
  projectId: string,
  files: readonly { objectId: string; size: number; status: string }[],
): {
  projectId: string;
  title: string;
  files: readonly { objectId: string; size: number; status: string }[];
} {
  try {
    validateProjectId(projectId);
  } catch {
    throw new Error("invalid file management");
  }
  if (files.some((file) => !file.objectId || !Number.isSafeInteger(file.size) || file.size < 0))
    throw new Error("invalid file management");
  return { projectId, title: "File management", files: files.map((file) => ({ ...file })) };
}
