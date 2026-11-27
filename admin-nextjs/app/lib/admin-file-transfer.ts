import { adminHeaders, clearAdminSession, readAdminSession, projectAdminURL, systemUIURL } from "./admin-api.js";
import type { AdminContribution } from "./admin-registry.js";
import { adminDownloadName } from "./admin-file-data.js";

function transferURL(contribution: AdminContribution, key: string, extra: Record<string, string> = {}) {
  return projectAdminURL(contribution.projectID, "files/blob") + "?" + new URLSearchParams({ version: contribution.version, key, ...extra }).toString();
}

// XHR exposes measured upload progress without reading a large File into JS.
export function uploadBinaryFile(token: string, contribution: AdminContribution, key: string, filename: string, contentType: string, file: File, signal: AbortSignal, progress: (percent: number | undefined) => void): Promise<unknown> {
  return new Promise((resolve, reject) => {
    if (signal.aborted) { reject(new Error("Upload cancelled")); return; }
    const xhr = new XMLHttpRequest();
    const abort = () => xhr.abort();
    signal.addEventListener("abort", abort, { once: true });
    const cleanup = () => { signal.removeEventListener("abort", abort); };
    xhr.open("POST", transferURL(contribution, key, { confirmed: "true", filename, contentType }));
    xhr.setRequestHeader("Authorization", "Bearer " + token);
    xhr.setRequestHeader("Content-Type", "application/octet-stream");
    xhr.upload.onprogress = event => progress(event.lengthComputable && event.total ? event.loaded / event.total * 100 : undefined);
    xhr.upload.onload = () => progress(undefined); // durable server commit is still pending
    xhr.onload = () => {
      cleanup();
      if (xhr.status === 401 && readAdminSession() === token) {
        clearAdminSession();
        window.location.replace(systemUIURL());
      }
      if (xhr.status < 200 || xhr.status >= 300) { let message = xhr.responseText || "Upload failed (" + xhr.status + ")"; try { message = JSON.parse(message).message || message; } catch {} reject(new Error(message)); return; }
      try { resolve(JSON.parse(xhr.responseText)); } catch (reason) { reject(reason); }
    };
    xhr.onerror = () => { cleanup(); reject(new Error("File transfer failed")); };
    xhr.onabort = () => { cleanup(); reject(new Error("Upload cancelled")); };
    progress(0);
    xhr.send(file);
  });
}

export async function downloadBinaryFile(contribution: AdminContribution, key: string, filename: string, signal: AbortSignal) {
  const token = readAdminSession();
  if (!token) throw new Error("Sign in to download");
  // The quiet marker prevents the global toast interceptor from cloning and
  // decoding the entire attachment as text.
  const response = await fetch(transferURL(contribution, key), { headers: adminHeaders(token, { "x-brain-quiet": "1" }), signal, cache: "no-store" });
  if (!response.ok) throw new Error(await response.text());
  const blob = await response.blob();
  if (signal.aborted) return;
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url; anchor.download = adminDownloadName(filename);
  document.body.appendChild(anchor); anchor.click(); anchor.remove();
  window.setTimeout(() => URL.revokeObjectURL(url), 1000);
}
