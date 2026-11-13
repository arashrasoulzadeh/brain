export class FileAccessReader {
  constructor(private readonly grants: ReadonlySet<string>) {}
  read(projectId: string, objectId: string, actorId: string): void {
    if (!this.grants.has(`${projectId}:${objectId}:${actorId}`)) throw new Error("file access denied");
  }
}
