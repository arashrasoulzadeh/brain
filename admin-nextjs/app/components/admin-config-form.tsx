"use client";
import { useState } from "react";
import { adminHeaders } from "../lib/admin-api";
import { configurationPatch } from "../lib/admin-config";

export function AdminConfigForm({ token, projectID, pluginID, version, schema }: {
  token: string; projectID: string; pluginID: string; version: string; schema: Record<string, string>;
}) {
  const [values, setValues] = useState<Record<string, string>>({});
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const fields = Object.entries(schema);
  if (!fields.length) return null;
  function change(key: string, value: string | undefined) {
    setValues(previous => {
      const next = { ...previous };
      if (value === undefined) delete next[key];
      else Object.defineProperty(next, key, { value, enumerable: true, writable: true, configurable: true });
      return next;
    });
  }
  async function submit() {
    setMessage("");
    try {
      const patch = configurationPatch(schema, values);
      if (!Object.keys(patch).length) return;
      if (!window.confirm("Update configuration for " + projectID + " / " + pluginID + "@" + version + "?")) return;
      setBusy(true);
      const response = await fetch("/api/admin/actions", { method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }),
        body: JSON.stringify({ operation: "plugin.configure", project_id: projectID, plugin_id: pluginID, version, values: patch, confirmed: true }) });
      if (!response.ok) throw new Error(await response.text());
      setValues({}); setMessage("Configuration updated");
    } catch (reason) { setMessage(String(reason)); }
    finally { setBusy(false); }
  }
  return <section className="section admin-action-panel"><h2>Configuration</h2><p className="muted">Only changed fields are submitted. Existing values are not exposed.</p>
    <form className="admin-operation-form" onSubmit={event => { event.preventDefault(); if (!busy) void submit(); }}>
      {fields.map(([name, type]) => <div key={name}><label>{name} ({type})
        {type === "bool" ? <select disabled={busy} value={Object.hasOwn(values, name) ? values[name] : ""} onChange={event => change(name, event.target.value || undefined)}>
          <option value="">Unchanged</option><option value="true">True</option><option value="false">False</option>
        </select> : <input disabled={busy || (type !== "int" && type !== "string")} type={type === "int" ? "number" : "text"} step={type === "int" ? 1 : undefined}
          placeholder="Unchanged" value={Object.hasOwn(values, name) ? values[name] : ""} onChange={event => change(name, event.target.value)} />}
        </label>{Object.hasOwn(values, name) && <button type="button" className="secondary" disabled={busy} onClick={() => change(name, undefined)}>Keep unchanged</button>}</div>)}
      <div className="operation-submit"><button type="submit" disabled={busy || !Object.keys(values).length}>{busy ? "Saving…" : "Save configuration"}</button></div>
    </form>{message && <p role="status">{message}</p>}
  </section>;
}
