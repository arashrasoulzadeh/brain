"use client";
import { useState } from "react";
import { adminHeaders, projectAdminURL, notifyPluginCatalogChanged } from "../lib/admin-api";
import { configurationPatch } from "../lib/admin-config";
import { AdminDialog, AdminProgress } from "./admin-dialog";
import { DEFAULT_FILE_LIMIT, MAX_FILE_MB, fileLimitBytes, fileLimitMB } from "../lib/admin-file-data";

export function AdminConfigForm({ token, projectID, pluginID, version, schema, configuredFields = [], publicConfig = {}, maxUploadBytes, onSaved }: {
  token: string; projectID: string; pluginID: string; version: string; schema: Record<string, string>; configuredFields?: string[]; publicConfig?: Record<string, string | number | boolean>; maxUploadBytes?: number; onSaved?: () => Promise<void>;
}) {
  const [open, setOpen] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [values, setValues] = useState<Record<string, string>>({});
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const fields = Object.entries(schema);
  if (!fields.length) return <section className="section admin-action-panel"><h2>Project settings</h2><p className="muted">This plugin declares no configurable fields.</p></section>;
  function change(key: string, value: string | undefined) {
    setValues(previous => {
      const next = { ...previous };
      if (value === undefined) delete next[key];
      else Object.defineProperty(next, key, { value, enumerable: true, writable: true, configurable: true });
      return next;
    });
  }
  async function submit(confirmed = false) {
    setMessage("");
    try {
      const input = { ...values };
      if (pluginID === "files" && Object.hasOwn(input, "maxUploadBytes")) input.maxUploadBytes = String(fileLimitBytes(input.maxUploadBytes));
      const patch = configurationPatch(schema, input);
      if (!Object.keys(patch).length) return;
      if (!confirmed) { setConfirming(true); return; }
      setConfirming(false);
      setBusy(true);
      const response = await fetch(projectAdminURL(projectID, "actions"), { method: "POST", headers: adminHeaders(token, { "content-type": "application/json" }),
        body: JSON.stringify({ operation: "plugin.configure", project_id: projectID, plugin_id: pluginID, version, values: patch, confirmed: true }) });
      if (!response.ok) throw new Error(await response.text());
      setValues({}); setMessage("Configuration updated");
      notifyPluginCatalogChanged(projectID);
      await onSaved?.();
    } catch (reason) { setMessage(String(reason)); }
    finally { setBusy(false); }
  }
  return <section className="section admin-action-panel"><div className="section-heading"><h2>Project settings</h2><button type="button" onClick={() => { setOpen(true); setMessage(""); setValues({}); }}>Edit settings</button></div><p className="muted">Settings apply only to {projectID} / {pluginID}@{version}. Only explicitly declared non-secret settings are readable. Other configured values can be replaced without revealing them.</p><dl>{fields.map(([name, type]) => <div key={name}><dt>{pluginID === "files" && name === "maxUploadBytes" ? "Maximum file size (MB)" : name + " (" + type + ")"}</dt><dd>{pluginID === "files" && name === "maxUploadBytes" ? fileLimitMB(maxUploadBytes ?? DEFAULT_FILE_LIMIT) + " MB" + (configuredFields.includes(name) ? "" : " (default)") : Object.hasOwn(publicConfig, name) ? String(publicConfig[name]) : configuredFields.includes(name) ? "Configured (hidden)" : "Not configured"}</dd></div>)}</dl>
    {open && <AdminDialog title="Edit project settings" scope={projectID + " · " + pluginID + "@" + version} busy={busy} onClose={() => { if (!busy) { setOpen(false); setConfirming(false); setValues({}); setMessage(""); } }}>
    <form className="admin-operation-form" onSubmit={event => { event.preventDefault(); if (!busy && !confirming) void submit(); }}>
      {fields.map(([name, type]) => <div key={name}><label>{pluginID === "files" && name === "maxUploadBytes" ? "Maximum file size (MB)" : name + " (" + type + ")"}
        {type === "bool" ? <select disabled={busy || confirming} value={Object.hasOwn(values, name) ? values[name] : ""} onChange={event => change(name, event.target.value || undefined)}>
          <option value="">Unchanged</option><option value="true">True</option><option value="false">False</option>
        </select> : <input disabled={busy || confirming || (type !== "int" && type !== "string")} type={type === "int" ? "number" : "text"} step={pluginID === "files" && name === "maxUploadBytes" ? 0.000001 : type === "int" ? 1 : undefined}
          min={pluginID === "files" && name === "maxUploadBytes" ? 0.000001 : undefined} max={pluginID === "files" && name === "maxUploadBytes" ? MAX_FILE_MB : undefined} placeholder={pluginID === "files" && name === "maxUploadBytes" ? fileLimitMB(maxUploadBytes ?? DEFAULT_FILE_LIMIT) : Object.hasOwn(publicConfig, name) ? String(publicConfig[name]) : "Unchanged"} value={Object.hasOwn(values, name) ? values[name] : ""} onChange={event => change(name, event.target.value)} />}
        </label>{pluginID === "files" && name === "maxUploadBytes" && <small className="muted">Maximum 1,000 MB (1 GB). 1 MB = 1,000,000 bytes.</small>}{Object.hasOwn(values, name) && <button type="button" className="secondary" disabled={busy || confirming} onClick={() => change(name, undefined)}>Keep unchanged</button>}</div>)}
      <div className="operation-submit"><button type="submit" disabled={busy || confirming || !Object.keys(values).length}>{busy ? "Saving…" : "Save configuration"}</button></div>
    </form>{confirming && <section className="admin-confirmation"><p>Save changes to this project installation only?</p><div className="lifecycle-actions"><button type="button" onClick={() => void submit(true)}>Confirm settings</button><button type="button" className="secondary" onClick={() => setConfirming(false)}>Cancel confirmation</button></div></section>}{busy && <AdminProgress label="Saving project settings…" />}{message && <p role="status">{message}</p>}
    </AdminDialog>}
  </section>;
}
