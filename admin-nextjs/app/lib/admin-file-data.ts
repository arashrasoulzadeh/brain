// Existing Files routes store strings. A deliberately identifiable data URI
// preserves selected file bytes without changing the existing text contract.
export const ADMIN_FILE_LIMIT = 40 * 1024; // legacy JSON transport only
export const FILE_MB = 1_000_000;
export const MAX_FILE_MB = 1000;
export const DEFAULT_FILE_LIMIT = 100 * FILE_MB;
export function fileLimitBytes(megabytes: string): number {
  const match = /^(\d+)(?:\.(\d{1,6}))?$/.exec(megabytes);
  if (!match) throw new Error("Enter a file limit in MB, with at most six decimal places");
  const bytes = Number(match[1]) * FILE_MB + Number((match[2] ?? "").padEnd(6, "0"));
  if (!Number.isSafeInteger(bytes) || bytes < 1 || bytes > MAX_FILE_MB * FILE_MB) throw new Error("File limit must be greater than zero and at most 1,000 MB (1 GB)");
  return bytes;
}
export function fileLimitMB(bytes: number): string { return String(bytes / FILE_MB); }
const prefix = "data:application/octet-stream;base64,";

export function encodeAdminFile(bytes: Uint8Array): string {
  if (bytes.byteLength > ADMIN_FILE_LIMIT) throw new Error("Choose a file of 40 KiB or less");
  return prefix + btoa(Array.from(bytes, byte => String.fromCharCode(byte)).join(""));
}

export function decodeAdminFile(data: string): Uint8Array | undefined {
  if (!data.startsWith(prefix)) return undefined;
  const encoded = data.slice(prefix.length);
  if (encoded.length > Math.ceil(ADMIN_FILE_LIMIT / 3) * 4 || encoded.length % 4 !== 0 ||
      !/^[A-Za-z0-9+/]*={0,2}$/.test(encoded)) return undefined;
  try {
    const binary = atob(encoded);
    if (btoa(binary) !== encoded || binary.length > ADMIN_FILE_LIMIT) return undefined;
    return Uint8Array.from(binary, char => char.charCodeAt(0));
  } catch {
    return undefined;
  }
}

export function adminDownloadName(key: string): string {
  const name = key.split(/[/\\]/).pop()?.replace(/[\x00-\x1f\x7f]/g, "").trim();
  return name && name !== "." && name !== ".." ? name : "download.bin";
}
