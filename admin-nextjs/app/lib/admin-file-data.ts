// Existing Files routes store strings. A deliberately identifiable data URI
// preserves selected file bytes without changing the existing text contract.
export const ADMIN_FILE_LIMIT = 40 * 1024;
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
