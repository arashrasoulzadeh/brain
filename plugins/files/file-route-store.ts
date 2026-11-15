type FileRecord = { key: string; data: string; published: boolean };

export class FileRouteStore {
  private readonly files = new Map<string, FileRecord>();

  execute(
    operation: "upload" | "download" | "metadata" | "publish",
    key: string,
    data?: string,
  ): unknown {
    const existing = this.files.get(key);
    if (operation === "upload") {
      if (!data) throw new Error("file data is required");
      const record = { key, data, published: false };
      this.files.set(key, record);
      return { key, uploaded: true, published: false };
    }
    if (!existing) throw new Error("file not found");
    if (operation === "publish") {
      existing.published = true;
      return { key, published: true };
    }
    if (operation === "metadata")
      return { key, published: existing.published, size: existing.data.length };
    return { key, data: existing.data, published: existing.published };
  }
}

export const fileRouteStore = new FileRouteStore();
