"use client";
import { useEffect, useRef, useState } from "react";
import type { AdminContribution } from "../lib/admin-registry";
import { downloadBinaryFile } from "../lib/admin-file-transfer";
import { ADMIN_FILE_LIMIT, DEFAULT_FILE_LIMIT, MAX_FILE_MB, FILE_MB, adminDownloadName, decodeAdminFile, encodeAdminFile } from "../lib/admin-file-data";

export function FileUploadField({ value, disabled, limit = DEFAULT_FILE_LIMIT, onData, onFile, onBinary, onReading, onProgress }: {
  value: string; disabled: boolean; limit?: number; onProgress: (percent: number | undefined) => void; onData: (value: string) => void;
  onFile: (name: string, data: string, contentType: string) => void; onBinary: (file: File) => void; onReading: (reading: boolean) => void;
}) {
  const [mode, setMode] = useState<"file" | "text">("file");
  const [selected, setSelected] = useState<{ name: string; size: number }>();
  const [error, setError] = useState("");
  const generation = useRef(0);
  const readerRef = useRef<FileReader | undefined>(undefined);
  const maximum = Math.min(MAX_FILE_MB * FILE_MB, limit);
  useEffect(() => () => { generation.current++; readerRef.current?.abort(); }, []);

  async function pick(file: File | undefined) {
    const current = ++generation.current;
    setError(""); setSelected(undefined); onData("");
    if (!file) return;
    if (file.size > maximum) { setError("Choose a file of " + maximum / FILE_MB + " MB or less."); return; }
    if (file.size > ADMIN_FILE_LIMIT) { onBinary(file); setSelected({ name: file.name, size: file.size }); return; }
    onProgress(0); onReading(true);
    try {
      const bytes = await new Promise<ArrayBuffer>((resolve, reject) => {
        const reader = new FileReader(); readerRef.current = reader;
        reader.onprogress = event => { if (current === generation.current) onProgress(event.lengthComputable && event.total ? event.loaded / event.total * 100 : undefined); };
        reader.onload = () => resolve(reader.result as ArrayBuffer);
        reader.onerror = () => reject(reader.error ?? new Error("Could not read file"));
        reader.onabort = () => reject(new Error("File reading cancelled"));
        reader.readAsArrayBuffer(file);
      });
      const data = encodeAdminFile(new Uint8Array(bytes));
      if (current !== generation.current) return;
      onFile(file.name, data, file.type);
      setSelected({ name: file.name, size: file.size });
    } catch (reason) {
      if (current === generation.current) setError(reason instanceof Error ? reason.message : String(reason));
    } finally {
      if (current === generation.current) { readerRef.current = undefined; onReading(false); onProgress(undefined); }
    }
  }

  return <fieldset className="wide-field admin-file-upload">
    <legend>File contents *</legend>
    <div className="lifecycle-actions">
      <button type="button" className="secondary" aria-pressed={mode === "file"} disabled={disabled} onClick={() => { if (mode !== "file") { setMode("file"); onData(""); setError(""); } }}>Choose a file</button>
      <button type="button" className="secondary" aria-pressed={mode === "text"} disabled={disabled} onClick={() => { if (mode !== "text") { setMode("text"); setSelected(undefined); onData(""); setError(""); } }}>Enter text</button>
    </div>
    {mode === "file" ? <>
      <label>Local file<input type="file" disabled={disabled} onChange={event => {
        const file = event.currentTarget.files?.[0];
        event.currentTarget.value = "";
        void pick(file);
      }} /></label>
      <p className="muted">Maximum {maximum / FILE_MB} MB for this project. Large files use binary transfer; small files retain the legacy JSON format.</p>
      {selected && value && <p role="status">{selected.name} · {selected.size / FILE_MB} MB · ready to upload</p>}
    </> : <label>data *<textarea rows={5} value={value} required disabled={disabled} onChange={event => onData(event.target.value)} /></label>}
    {error && <p role="alert">{error}</p>}
  </fieldset>;
}

export function FileDownloadResult({ result, contribution }: { result: unknown; contribution: AdminContribution }) {
  const [downloading, setDownloading] = useState(false);
  const [error, setError] = useState("");
  const active = useRef<AbortController | null>(null);
  useEffect(() => () => active.current?.abort(), []);
  async function download(key: string, filename: string) {
    if (active.current) return;
    const controller = new AbortController(); active.current = controller;
    setDownloading(true); setError("");
    try { await downloadBinaryFile(contribution, key, filename, controller.signal); }
    catch (reason) { if (!controller.signal.aborted) setError(String(reason)); }
    finally { active.current = null; if (!controller.signal.aborted) setDownloading(false); }
  }
  if (result && typeof result === "object" && "streamed" in result && result.streamed === true && "key" in result && typeof result.key === "string") {
    const key = result.key;
    const meta = "metadata" in result && result.metadata && typeof result.metadata === "object" ? result.metadata : {};
    const filename = "filename" in meta && typeof meta.filename === "string" ? meta.filename : key;
    return <section className="section"><h3>File ready to download</h3><p>{filename}</p><button type="button" disabled={downloading} onClick={() => void download(key, filename)}>{downloading ? "Downloading…" : "Save original file"}</button>{downloading && <progress aria-label="Downloading file" />}{error && <p role="alert">{error}</p>}</section>;
  }
  if (!result || typeof result !== "object" || !("key" in result) || !("data" in result) ||
      typeof result.key !== "string" || typeof result.data !== "string") return <p role="alert">Invalid file response</p>;
  const key = result.key, data = result.data;
  const metadata = "metadata" in result && result.metadata && typeof result.metadata === "object" ? result.metadata : undefined;
  const filename = metadata && "filename" in metadata && typeof metadata.filename === "string" ? metadata.filename : key;
  const decoded = decodeAdminFile(data);
  function save(bytes: Uint8Array) {
    // Never open plugin-controlled data as a page or executable MIME type.
    const url = URL.createObjectURL(new Blob([new Uint8Array(bytes).buffer], { type: "application/octet-stream" }));
    const anchor = document.createElement("a");
    anchor.href = url; anchor.download = adminDownloadName(filename);
    document.body.appendChild(anchor);
    anchor.click(); anchor.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
  return <section className="section">
    <h3>File ready to download</h3>
    <p>{filename} · {decoded ? decoded.byteLength : new TextEncoder().encode(data).byteLength} bytes</p>
    <div className="lifecycle-actions">
      {decoded && <button type="button" onClick={() => save(decoded)}>Save original file</button>}
      <button type="button" className={decoded ? "secondary" : undefined} onClick={() => save(new TextEncoder().encode(data))}>{decoded ? "Save stored text" : "Save file"}</button>
    </div>
    {decoded && <p className="muted">The original bytes are decoded from the stored base64 representation.</p>}
    {metadata && <dl>{Object.entries(metadata).map(([field, value]) => <div key={field}><dt>{field.replaceAll("_", " ")}</dt><dd>{String(value)}</dd></div>)}</dl>}
  </section>;
}
