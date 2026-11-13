import { validateProjectId } from "../sdk-typescript/src/index.js";

export function projectRoomID(projectId: string, roomId: string): string {
  validateProjectId(projectId);
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(roomId)) throw new Error("invalid project room id");
  return `${projectId}:${roomId}`;
}
