"use client";

import { useEffect, useRef, useState } from "react";
import { readAdminSession } from "../lib/admin-api";
import type { AdminContribution } from "../lib/admin-registry";
import type { AdminAction } from "../lib/admin-types";
import { invokeAdminAction } from "../lib/admin-sdk";

type ModelRecord = { id: string; revision: string; state?: string; fields: Record<string, unknown> };

function isRecord(value: unknown): value is ModelRecord {
  return !!value && typeof value === "object" && "id" in value && typeof value.id === "string" &&
    "revision" in value && typeof value.revision === "string" && "fields" in value &&
    !!value.fields && typeof value.fields === "object" && !Array.isArray(value.fields);
}

export function ModelPanel({ contribution }: { contribution: AdminContribution }) {
  const { name, schema: model } = contribution.page.model!;
  const [id, setId] = useState("");
  const [revision, setRevision] = useState("");
  const [fields, setFields] = useState<Record<string, string>>({});
  const [state, setState] = useState("");
  const [records, setRecords] = useState<ModelRecord[]>([]);
  const [cursors, setCursors] = useState([""]);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const pending = useRef<AbortController | null>(null);
  const modelRoutes = contribution.page.actions.filter(route => route.model_action);
  const listRoute = modelRoutes.find(route => route.model_action === "list");

  function choose(record: ModelRecord) {
    setId(record.id);
    setRevision(record.revision);
    setState(record.state ?? "");
    setFields(Object.fromEntries(Object.entries(record.fields).map(([key, value]) => [key, String(value)])));
  }

  useEffect(() => () => { pending.current?.abort(); pending.current = null; }, []);
  useEffect(() => {
    if (!listRoute) return;
    const controller = new AbortController();
    pending.current = controller;
    setBusy(true);
    void invokeAdminAction(readAdminSession(), contribution, listRoute.operation, {}, controller.signal, true).then(data => {
      if (!Array.isArray(data) || !data.every(isRecord)) throw new Error("Invalid model list");
      if (!controller.signal.aborted) setRecords(data);
    }).catch(error => { if (!controller.signal.aborted) { setRecords([]); setMessage(String(error)); } }).finally(() => {
      if (pending.current === controller) { pending.current = null; if (!controller.signal.aborted) setBusy(false); }
    });
    return () => { controller.abort(); pending.current?.abort(); pending.current = null; };
  }, [contribution]);

  async function invoke(route: AdminAction, nextCursors = cursors) {
    if (pending.current || !route.model_action) return;
    const action = route.model_action;
    if (!route.read_only && !window.confirm("Run " + route.title + " for " + contribution.projectID + " / " + name + " / " + id + "?")) return;
    const controller = new AbortController();
    pending.current = controller;
    setBusy(true); setMessage("");
    try {
      const payload: Record<string, unknown> = action === "list" ? { after: nextCursors[nextCursors.length - 1] } : { id };
      if (["update", "delete", "transition"].includes(action)) payload.revision = revision;
      if (["create", "update"].includes(action)) {
        payload.fields = Object.fromEntries(Object.entries(model.fields).map(([key, kind]) => {
          const value = fields[key] ?? (kind === "boolean" ? "false" : "");
          if (kind === "integer" && (!/^-?\d+$/.test(value) || !Number.isSafeInteger(Number(value)))) throw new Error("Invalid integer: " + key);
          return [key, kind === "integer" ? Number(value) : kind === "boolean" ? value === "true" : value];
        }));
      }
      const data = await invokeAdminAction(readAdminSession(), contribution, route.operation, payload, controller.signal, Boolean(route.read_only));
      if (controller.signal.aborted) return;
      if (action === "list") {
        if (!Array.isArray(data) || !data.every(isRecord)) throw new Error("Invalid model list");
        setRecords(data); setCursors(nextCursors);
      } else if (action === "delete") {
        setId(""); setRevision(""); setState(""); setFields({}); setRecords(rows => rows.filter(row => row.id !== id));
      } else {
        if (!isRecord(data)) throw new Error("Invalid model record");
        choose(data);
        setRecords(rows => rows.map(row => row.id === data.id ? data : row));
      }
      setMessage(route.title + " completed");
    } catch (error) {
      if (!controller.signal.aborted) { setMessage(String(error)); if (action === "list") setRecords([]); }
    } finally {
      if (pending.current === controller) { pending.current = null; if (!controller.signal.aborted) setBusy(false); }
    }
  }

  return <section className="section admin-action-panel">
    <div className="panel-heading"><h2>{name}</h2><p className="muted">{contribution.projectID} · {contribution.pluginID}@{contribution.version}</p></div>
    <p className="muted">Schema v{model.version} · {state || model.initial || "No state machine"}</p>
    <div className="playground model-form">
      <label>Record ID<input disabled={busy} value={id} onChange={event => { setId(event.target.value); setRevision(""); setState(""); setFields({}); }} /></label>
      {Object.entries(model.fields).map(([key, kind]) => <label key={key}>{key}
        {kind === "boolean" ? <select disabled={busy} value={fields[key] ?? "false"} onChange={event => setFields(values => ({ ...values, [key]: event.target.value }))}><option>false</option><option>true</option></select> :
          <input disabled={busy} type={kind === "integer" ? "number" : "text"} step={kind === "integer" ? 1 : undefined} value={fields[key] ?? ""} onChange={event => setFields(values => ({ ...values, [key]: event.target.value }))} />}
      </label>)}
      <div className="playground-actions">{modelRoutes.filter(route => route.model_action !== "list").map(route => {
        const transition = route.model_transition ? model.transitions?.[route.model_transition] : undefined;
        const needsRevision = ["update", "delete", "transition"].includes(route.model_action!);
        return <button className={route.model_action === "delete" ? "danger" : route.read_only ? "secondary" : undefined} key={route.operation} type="button" disabled={busy || !id || (needsRevision && !revision) || (!!transition && state !== transition.from)} onClick={() => void invoke(route)}>{route.title}</button>;
      })}</div>
      {message && <p role="status">{message}</p>}
      {listRoute && <section className="admin-inventory-panel"><div className="section-heading"><h3>Records</h3><button className="secondary" type="button" disabled={busy} onClick={() => void invoke(listRoute, [""])}>Refresh</button></div>
        <div className="table-wrap"><table><thead><tr><th>ID</th><th>State</th>{Object.keys(model.fields).map(key => <th key={key}>{key}</th>)}<th>Action</th></tr></thead>
          <tbody>{records.map(record => <tr key={record.id}><td>{record.id}</td><td>{record.state ?? "—"}</td>{Object.keys(model.fields).map(key => <td key={key}>{String(record.fields[key] ?? "")}</td>)}<td><button className="secondary" type="button" disabled={busy} onClick={() => choose(record)}>Select</button></td></tr>)}</tbody></table></div>
        {!records.length && <p>{busy ? "Loading…" : "No records on this page."}</p>}
        <div className="lifecycle-actions"><button type="button" disabled={busy || cursors.length < 2} onClick={() => void invoke(listRoute, cursors.slice(0, -1))}>Previous</button><span>Page {cursors.length}</span>
          <button type="button" disabled={busy || records.length !== 20} onClick={() => void invoke(listRoute, [...cursors, records[records.length - 1].id])}>Next</button></div>
      </section>}
    </div>
  </section>;
}
