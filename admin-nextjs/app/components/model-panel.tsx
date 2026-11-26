"use client";

import { useState } from "react";
import { adminHeaders } from "../lib/admin-api";

export type DeclaredModel = { version: number; fields: Record<string, "string" | "integer" | "boolean">; initial?: string; transitions?: Record<string, { from: string; to: string }> };
export type DeclaredModelRoute = { Operation: string; Capability: string; Helper?: string; Model?: { name: string; action: string; transition?: string } };
type ModelRecord = { id: string; revision: string; state?: string; fields: Record<string, unknown> };

export function ModelPanel({ name, model, routes, session, projectId, pluginId, version }: {
  name: string;
  model: DeclaredModel;
  routes: DeclaredModelRoute[];
  session: string;
  projectId: string;
  pluginId: string;
  version: string;
}) {
  const [id, setId] = useState("");
  const [revision, setRevision] = useState("");
  const [fields, setFields] = useState<Record<string, string>>({});
  const [state, setState] = useState("");
  const [records, setRecords] = useState<ModelRecord[]>([]);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);

  function choose(record: ModelRecord) {
    setId(record.id);
    setRevision(record.revision);
    setState(record.state ?? "");
    setFields(Object.fromEntries(Object.entries(record.fields).map(([key, value]) => [key, String(value)])));
  }

  async function invoke(route: DeclaredModelRoute, after = "") {
    if (busy || !route.Model) return;
    const action = route.Model.action;
    if (action === "delete" && !window.confirm("Delete this record?")) return;
    setBusy(true);
    try {
      const payload: Record<string, unknown> = action === "list" ? { after } : { id };
      if (["update", "delete", "transition"].includes(action)) payload.revision = revision;
      if (["create", "update"].includes(action)) {
        payload.fields = Object.fromEntries(Object.entries(model.fields).map(([key, kind]) => {
          const value = fields[key] ?? (kind === "boolean" ? "false" : "");
          if (kind === "integer" && (!value.trim() || !Number.isSafeInteger(Number(value)))) throw new Error("Invalid integer: " + key);
          return [key, kind === "integer" ? Number(value) : kind === "boolean" ? value === "true" : value];
        }));
      }
      const response = await fetch("/api/admin/actions", {
        method: "POST",
        headers: adminHeaders(session, { "content-type": "application/json" }),
        body: JSON.stringify({ operation: "plugin.invoke", project_id: projectId, plugin_id: pluginId, version, target_operation: route.Operation, payload }),
      });
      const result = await response.json() as { Success?: boolean; Payload?: string; ErrorCode?: string; message?: string };
      if (!response.ok || !result.Success || !result.Payload) throw new Error(result.message || result.ErrorCode || "Operation failed");
      const data: unknown = JSON.parse(new TextDecoder().decode(Uint8Array.from(atob(result.Payload), (char) => char.charCodeAt(0))));
      if (action === "list") setRecords(data as ModelRecord[]);
      else if (action === "delete") { setId(""); setRevision(""); setState(""); setFields({}); setRecords((rows) => rows.filter((row) => row.id !== id)); }
      else choose(data as ModelRecord);
      setMessage(route.Operation + " completed");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Operation failed");
    } finally {
      setBusy(false);
    }
  }

  const modelRoutes = routes.filter((route) => route.Helper === "host.model" && route.Model?.name === name);
  const listRoute = modelRoutes.find((route) => route.Model?.action === "list");
  return <section className="section">
    <h2>{name}</h2><p className="muted">Schema v{model.version} · {state || model.initial || "No state machine"}</p>
    <div className="playground">
      <label>Record ID<input value={id} onChange={(event) => { setId(event.target.value); setRevision(""); setState(""); }} /></label>
      {Object.entries(model.fields).map(([key, kind]) => <label key={key}>{key}
        {kind === "boolean" ? <select value={fields[key] ?? "false"} onChange={(event) => setFields((values) => ({ ...values, [key]: event.target.value }))}><option>false</option><option>true</option></select> :
          <input type={kind === "integer" ? "number" : "text"} value={fields[key] ?? ""} onChange={(event) => setFields((values) => ({ ...values, [key]: event.target.value }))} />}
      </label>)}
      <div className="playground-actions">{modelRoutes.map((route) => <button key={route.Operation} disabled={busy || (["update", "delete", "transition"].includes(route.Model!.action) && !revision)} onClick={() => void invoke(route)}>{route.Model!.transition ?? route.Model!.action}</button>)}</div>
      {message && <p role="status">{message}</p>}
      {records.map((record) => <button className="secondary" key={record.id} onClick={() => choose(record)}>{record.id} {record.state}</button>)}
      {listRoute && records.length === 20 && <button disabled={busy} onClick={() => void invoke(listRoute, records[records.length - 1].id)}>Next page</button>}
    </div>
  </section>;
}
