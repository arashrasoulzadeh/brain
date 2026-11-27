"use client";
import { useRef, useState } from "react";
import { AdminDialog, AdminProgress } from "./admin-dialog";
import type { PluginManifest } from "../lib/admin-types";

export function PluginInstallDialog({ projectID, manifest, onInstall, onClose }: {
  projectID: string; manifest: PluginManifest; onInstall: () => Promise<void>; onClose: () => void;
}) {
  const running = useRef(false);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");
  async function install() {
    if (running.current || done) return;
    running.current = true; setBusy(true); setError("");
    try { await onInstall(); setDone(true); }
    catch (reason) { setError(reason instanceof Error ? reason.message : String(reason)); }
    finally { running.current = false; setBusy(false); }
  }
  return <AdminDialog title={"Install " + (manifest.Title || manifest.ID)} scope={projectID + " · " + manifest.ID + "@" + manifest.Version} busy={busy} onClose={onClose}>
    <p>Installing will grant all capabilities listed below within this project, enable the plugin and start its worker.</p>
    <p className="muted">These are project-wide plugin grants; resource access policies still apply. Other projects do not receive grants. Successful worker startup marks the shared artifact ready; existing active versions are not switched.</p>
    <h3>Requested capabilities ({manifest.Capabilities?.length ?? 0})</h3>
    {manifest.Capabilities?.length ? <ul>{manifest.Capabilities.map(capability => <li key={capability}><code>{capability}</code></li>)}</ul> : <p>No capabilities requested.</p>}
    {!!manifest.Dependencies?.length && <><h3>Required dependencies</h3><ul>{manifest.Dependencies.map(dep => <li key={dep}><code>{dep}</code></li>)}</ul><p>These must already be enabled in this project.</p></>}
    {!!manifest.OptionalDependencies?.length && <><h3>Optional dependencies</h3><ul>{manifest.OptionalDependencies.map(dep => <li key={dep}><code>{dep}</code></li>)}</ul><p className="muted">Optional dependencies do not block installation and are not automatically enabled.</p></>}
    {manifest.ID === "access-management" && <p role="note">Access management enforces explicit resource policies. Existing plugin requests may be denied until rules are configured.</p>}
    {Object.keys(manifest.ConfigSchema ?? {}).length > 0 && <p>After installation, use Manage to configure this project's settings.</p>}
    {busy && <AdminProgress label="Installing, granting capabilities and starting worker…" />}
    {done && <p role="status">Plugin installed and enabled.</p>}
    {error && <p role="alert">{error}</p>}
    <div className="lifecycle-actions">
      {!done && <button type="button" disabled={busy} onClick={() => void install()}>Install and enable</button>}
      <button type="button" className="secondary" disabled={busy} onClick={onClose}>{done ? "Done" : "Cancel"}</button>
    </div>
  </AdminDialog>;
}
