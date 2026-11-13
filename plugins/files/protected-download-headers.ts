export function protectedDownloadHeaders(filename: string): Record<string, string> {
  if (!filename || filename.includes("\r") || filename.includes("\n")) throw new Error("invalid download filename");
  return { "content-disposition": `attachment; filename="${filename.replaceAll('"', "")}"`, "cache-control": "private, no-store" };
}
