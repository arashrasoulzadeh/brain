"use client";
import { useEffect, useRef, useState } from "react";
import { ADMIN_FILE_LIMIT, adminDownloadName, decodeAdminFile, encodeAdminFile } from "../lib/admin-file-data";

export function FileUploadField({ value, disabled, onData, onFile, onReading }: {
  value: string; disabled: boolean; onData: (value: string) => void;
  onFile: (name: string, data: string) => void; onReading: (reading: boolean) => void;
}) {
  const [mode, setMode] = useState<"file" | "text">("file");
  const [selected, setSelected] = useState<{ name: string; size: number }>();
  const [error, setError] = useState("");
  const generation = useRef(0);
  useEffect(() => () => { generation.current++; }, []);

  async function pick(file: File | undefined) {
    const current = ++generation.current;
    setError(""); setSelected(undefined); onData("");
    if (!file) return;
    if (file.size > ADMIN_FILE_LIMIT) { setError("Choose a file of 40 KiB or less."); return; }
    onReading(true);
    try {
      const data = encodeAdminFile(new Uint8Array(await file.arrayBuffer()));
      if (current !== generation.current) return;
      onFile(file.name, data);
      setSelected({ name: file.name, size: file.size });
    } catch (reason) {
      if (current === generation.current) setError(reason instanceof Error ? reason.message : String(reason));
    } finally {
      if (current === generation.current) onReading(false);
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
      <p className="muted">Maximum 40 KiB. Files are stored as lossless base64 text; use Download file to save the original bytes.</p>
      {selected && value && <p role="status">{selected.name} · {selected.size} bytes · ready to upload</p>}
    </> : <label>data *<textarea rows={5} value={value} required disabled={disabled} onChange={event => onData(event.target.value)} /></label>}
    {error && <p role="alert">{error}</p>}
  </fieldset>;
}

export function FileDownloadResult({ result }: { result: unknown }) {
  if (!result || typeof result !== "object" || !("key" in result) || !("data" in result) ||
      typeof result.key !== "string" || typeof result.data !== "string") return <p role="alert">Invalid file response</p>;
  const key = result.key, data = result.data;
  const decoded = decodeAdminFile(data);
  function save(bytes: Uint8Array) {
    // Never open plugin-controlled data as a page or executable MIME type.
    const url = URL.createObjectURL(new Blob([new Uint8Array(bytes).buffer], { type: "application/octet-stream" }));
    const anchor = document.createElement("a");
    anchor.href = url; anchor.download = adminDownloadName(key);
    document.body.appendChild(anchor);
    anchor.click(); anchor.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
  return <section className="section">
    <h3>File ready to download</h3>
    <p>{key} · {decoded ? decoded.byteLength : new TextEncoder().encode(data).byteLength} bytes</p>
    <div className="lifecycle-actions">
      {decoded && <button type="button" onClick={() => save(decoded)}>Save original file</button>}
      <button type="button" className={decoded ? "secondary" : undefined} onClick={() => save(new TextEncoder().encode(data))}>{decoded ? "Save stored text" : "Save file"}</button>
    </div>
    {decoded && <p className="muted">The original bytes are decoded from the stored base64 representation.</p>}
  </section>;
}
