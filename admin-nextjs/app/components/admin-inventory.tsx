"use client";
import { useEffect, useState } from "react";
import { readAdminSession } from "../lib/admin-api";
import type { AdminContribution } from "../lib/admin-registry";
import type { AdminInventoryPage } from "../lib/admin-types";
import { adminPageRevision, fetchAdminInventory } from "../lib/admin-sdk";

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
    let revision: number | undefined;
    setLoading(true); setError(""); setPage(undefined);
    async function read() {
      const token = readAdminSession();
      if (!token) throw new Error("Sign in to continue");
      if (contribution.page.refresh_on?.length) revision = await adminPageRevision(token, contribution, signal);
      const result = await fetchAdminInventory(token, contribution, after, signal);
      if (!signal.aborted) { setPage(result); setLoading(false); }
    }
    function failed(reason: unknown) {
      if (!signal.aborted) { setPage(undefined); setError(String(reason)); setLoading(false); }
    }
    async function poll() {
      if (signal.aborted) return;
      try {
        if (!document.hidden) {
          const next = await adminPageRevision(readAdminSession(), contribution, signal);
          if (next !== revision && !signal.aborted) await read();
        }
      } catch (reason) { failed(reason); return; }
      if (!signal.aborted) timer = setTimeout(() => void poll(), 5000);
    }
    void read().then(() => {
      if (!signal.aborted && contribution.page.refresh_on?.length) timer = setTimeout(() => void poll(), 5000);
    }).catch(failed);
    return () => { controller.abort(); clearTimeout(timer); };
  }, [contribution, spec, after, refresh, reload]);
  if (!spec) return null;
  const fileColumns = spec.kind === "file_keys";
  return <section className="admin-inventory-panel">
    <div className="section-heading"><h3>{spec.title}</h3>
      <button type="button" className="secondary" disabled={loading || disabled} onClick={() => setReload(value => value + 1)}>Refresh</button></div>
    {error && <p role="alert">{error}</p>}
    {loading && <p role="status">Loading resources…</p>}
    {page && <><div className="table-wrap"><table><thead><tr><th>Resource</th>{fileColumns && <><th>Bytes</th><th>Published</th></>}{onSelect && <th>Action</th>}</tr></thead>
      <tbody>{page.rows.map(row => <tr key={row.key}><td>{row.key}</td>{fileColumns && <><td>{row.size_bytes ?? "—"}</td><td>{row.published === undefined ? "—" : row.published ? "Yes" : "No"}</td></>}{onSelect && <td><button className="secondary" type="button" disabled={disabled} onClick={() => onSelect(row.key)}>Select</button></td>}</tr>)}</tbody>
    </table></div>
    {page.rows.length === 0 && <p className="muted empty-state">No resources on this page.</p>}
    <div className="lifecycle-actions">
      <button type="button" disabled={history.length < 2 || loading} onClick={() => setHistory(previous => previous.slice(0, -1))}>Previous</button>
      <span>Page {history.length}</span>
      <button type="button" disabled={!page.next || loading} onClick={() => { if (page.next) setHistory(previous => [...previous, page.next!]); }}>Next</button>
    </div></>}
  </section>;
}
