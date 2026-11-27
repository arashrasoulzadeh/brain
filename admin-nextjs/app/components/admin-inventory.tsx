"use client";
import { useEffect, useState } from "react";
import { readAdminSession } from "../lib/admin-api";
import type { AdminContribution } from "../lib/admin-registry";
import type { AdminInventoryPage } from "../lib/admin-types";
import { fetchAdminInventory } from "../lib/admin-sdk";

export function AdminInventory({ contribution, refresh, disabled, onSelect }: {
  contribution: AdminContribution; refresh: number; disabled: boolean; onSelect?: (key: string) => void;
}) {
  const [history, setHistory] = useState([""]);
  const [page, setPage] = useState<AdminInventoryPage>();
  const [error, setError] = useState("");
  const [reload, setReload] = useState(0);
  const [loading, setLoading] = useState(true);
  const after = history[history.length - 1];
  const spec = contribution.page.inventory;
  useEffect(() => {
    if (!spec) return;
    const controller = new AbortController();
    const signal = controller.signal;
    let timer: ReturnType<typeof setTimeout>;
    setLoading(true); setError(""); setPage(undefined);
    async function read(quiet = false) {
      const token = readAdminSession();
      if (!token) throw new Error("Sign in to continue");
      const result = await fetchAdminInventory(token, contribution, after, signal, quiet);
      if (!signal.aborted) { setPage(result); setError(""); setLoading(false); }
    }
    function failed(reason: unknown) {
      if (!signal.aborted) { setPage(undefined); setError(String(reason)); setLoading(false); }
    }
    async function poll() {
      if (signal.aborted) return;
      try {
        // Policy and storage changes may not publish an observation. Re-read
        // the bounded, filtered page itself rather than retaining stale rows.
        if (!document.hidden) await read(true);
      } catch (reason) { failed(reason); }
      if (!signal.aborted) timer = setTimeout(() => void poll(), 5000);
    }
    void read().catch(failed).finally(() => {
      if (!signal.aborted) timer = setTimeout(() => void poll(), 5000);
    });
    return () => { controller.abort(); clearTimeout(timer); };
  }, [contribution, spec, after, refresh, reload]);
  if (!spec) return null;
  const fileColumns = spec.kind === "file_keys";
  return <section className="admin-inventory-panel">
    <div className="section-heading"><h3>{spec.title}</h3>
      <button type="button" className="secondary" disabled={loading || disabled} onClick={() => setReload(value => value + 1)}>Refresh</button></div>
    {error && <p role="alert">{error}</p>}
    {loading && <p role="status">Loading resources…</p>}
    {page && <><div className="table-wrap"><table><thead><tr><th>Resource</th>{fileColumns && <><th>Filename</th><th>Extension</th><th>Content type</th><th>Bytes</th><th>Updated (UTC)</th><th>Published</th></>}{onSelect && <th>Action</th>}</tr></thead>
      <tbody>{page.rows.map(row => <tr key={row.key}><td>{row.key}</td>{fileColumns && <><td>{row.filename ?? "—"}</td><td>{row.extension || "—"}</td><td>{row.content_type ?? "—"}</td><td>{row.size_bytes ?? "—"}{row.stored_bytes === undefined && <small className="muted"> stored</small>}</td><td>{row.updated_at ?? "—"}</td><td>{row.published === undefined ? "—" : row.published ? "Yes" : "No"}</td></>}{onSelect && <td><button className="secondary" type="button" disabled={disabled} onClick={() => onSelect(row.key)}>Select</button></td>}</tr>)}</tbody>
    </table></div>
    {page.rows.length === 0 && <p className="muted empty-state">No resources on this page.</p>}</>}
    {(page || history.length > 1) && <div className="lifecycle-actions">
      <button type="button" disabled={history.length < 2 || loading} onClick={() => setHistory(previous => previous.slice(0, -1))}>Previous</button>
      <span>Page {history.length}</span>
      <button type="button" disabled={!page?.next || loading} onClick={() => { if (page?.next) setHistory(previous => [...previous, page.next!]); }}>Next</button>
    </div>}
  </section>;
}
