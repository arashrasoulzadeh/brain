export class IdentityLinker {
  private readonly links = new Map<string, string>();
  link(userId: string, provider: string, subject: string): void {
    if (!userId.trim() || !provider.trim() || !subject.trim())
      throw new Error("invalid identity link");
    const key = `${provider}:${subject}`;
    const existing = this.links.get(key);
    if (existing && existing !== userId) throw new Error("identity already linked");
    this.links.set(key, userId);
  }
}
