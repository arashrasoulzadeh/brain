export class AuthorizedUploadWriter {
  constructor(
    private readonly grants: ReadonlySet<string>,
    private readonly bytes: Map<string, number>,
  ) {}
  write(projectId: string, objectId: string, actorId: string, chunk: Uint8Array): number {
    if (!this.grants.has(`${projectId}:${objectId}:${actorId}`)) throw new Error("upload denied");
    const key = `${projectId}:${objectId}`;
    const total = (this.bytes.get(key) ?? 0) + chunk.byteLength;
    this.bytes.set(key, total);
    return total;
  }
}
