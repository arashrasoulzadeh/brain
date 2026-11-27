"use client";
import { useState } from "react";
import type { AdminCopy } from "../lib/admin-copy";
import type { AuditRecord } from "../lib/admin-types";
import { filterAudit, type AuditOutcome } from "../lib/admin-observability";

const PAGE_SIZE = 50;
export function AuditPanel({ activePage, text, audit, auditFilter, setAuditFilter }: {
  activePage: string; text: AdminCopy; audit: AuditRecord[]; auditFilter: string; setAuditFilter: (value: string) => void;
}) {
  const [outcome, setOutcome] = useState<AuditOutcome>("all");
  const [source, setSource] = useState("");
  const [page, setPage] = useState(0);
  const records = filterAudit(audit, auditFilter, outcome, source);
  const sources = [...new Set(audit.map(record => record.source))].sort();
  const last = Math.max(0, Math.ceil(records.length / PAGE_SIZE) - 1);
  const current = Math.min(page, last);
  function exportRecords() {
    const url = URL.createObjectURL(new Blob([JSON.stringify(records, null, 2)], { type: "application/json" }));
    const anchor = document.createElement("a");
    anchor.href = url; anchor.download = "brain-audit.json";
    document.body.append(anchor);
    try { anchor.click(); } finally { anchor.remove(); window.setTimeout(() => URL.revokeObjectURL(url), 1000); }
  }
  return <section className={`section ${activePage !== "audit" ? "page-hidden" : ""}`} id="audit">
    <div className="section-heading"><div><p className="eyebrow">{text.audit}</p><h2>{text.audit}</h2></div><span className="count">{records.length} / {audit.length}</span></div>
    <div className="lifecycle-actions">
      <input className="plugin-search" aria-label={text.searchAudit} placeholder={text.searchAudit} value={auditFilter} onChange={event => { setAuditFilter(event.target.value); setPage(0); }} />
      <label>Outcome <select value={outcome} onChange={event => { setOutcome(event.target.value as AuditOutcome); setPage(0); }}><option value="all">All outcomes</option><option value="success">Success</option><option value="failure">Failed</option></select></label>
      <label>Source <select value={source} onChange={event => { setSource(event.target.value); setPage(0); }}><option value="">All sources</option>{sources.map(value => <option key={value} value={value}>{value}</option>)}</select></label>
      <button type="button" className="secondary" disabled={!records.length} onClick={exportRecords}>Export filtered JSON</button>
    </div>
    <p className="muted">Current server audit history only. Exported records may contain operational metadata.</p>
    <div className="audit-list">{records.slice(current * PAGE_SIZE, (current + 1) * PAGE_SIZE).map((record, index) => <div className="audit" key={record.action + "-" + index}>
      <span className={record.success ? "success" : "failure"}>{record.success ? "✓" : "!"}</span>
      <div><strong>{record.action}</strong><small>{record.source}{record.plugin_id ? " · " + record.plugin_id + "@" + record.version : ""}{record.project_id ? " · " + record.project_id : ""}{record.actor_id ? " · " + record.actor_id : ""}{record.success ? " · Success" : " · Failed"}</small>
      {record.occurred_at && <small><time dateTime={record.occurred_at}>{record.occurred_at}</time></small>}{record.detail && <p className="muted">{record.detail}</p>}</div>
    </div>)}{!records.length && <p className="muted">{audit.length ? "No audit events match the filters." : text.noAudit}</p>}</div>
    {records.length > PAGE_SIZE && <div className="lifecycle-actions"><button type="button" disabled={current === 0} onClick={() => setPage(current - 1)}>Previous</button><span>Page {current + 1} of {last + 1}</span><button type="button" disabled={current === last} onClick={() => setPage(current + 1)}>Next</button></div>}
  </section>;
}
