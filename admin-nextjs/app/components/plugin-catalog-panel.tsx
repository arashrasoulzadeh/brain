import Link from "next/link";
import { useState } from "react";
import { PluginInstallDialog } from "./plugin-install-dialog";
import { projectUIURL } from "../lib/admin-api";
import type { AdminCopy } from "../lib/admin-copy";
import type { PluginManifest, PluginStatus } from "../lib/admin-types";

export function PluginCatalogPanel({ activePage, projectID, text, manifests, installations, pluginFilter, setPluginFilter, installPlugin, runPluginAction }: {
  activePage: string;
  projectID: string;
  text: AdminCopy;
  manifests: PluginManifest[];
  installations: PluginStatus[];
  pluginFilter: string;
  setPluginFilter: (value: string) => void;
  installPlugin: (manifest: PluginManifest, projectID: string) => Promise<void>;
  runPluginAction: (operation: string, installation: PluginStatus) => void;
}) {
  const [pending, setPending] = useState<{ manifest: PluginManifest; projectID: string }>();
  if (activePage !== "plugins" || !projectID) return null;
  const query = pluginFilter.trim().toLowerCase();
  const installed = new Map(installations.map(item => [item.PluginID + "@" + item.Version, item]));
  const rows = manifests.filter(manifest =>
    [manifest.ID, manifest.Title, manifest.Description, manifest.Version, ...(manifest.Capabilities ?? [])].join(" ").toLowerCase().includes(query));
  return <section className="section">
    {pending && pending.projectID === projectID && <PluginInstallDialog key={pending.projectID + ":" + pending.manifest.ID + "@" + pending.manifest.Version} manifest={pending.manifest} projectID={pending.projectID} onInstall={() => installPlugin(pending.manifest, pending.projectID)} onClose={() => setPending(undefined)} />}
    <div className="section-heading">
      <div><h2>{text.registered}</h2><p className="muted">{manifests.length} registered · {installations.length} installed · {installations.filter(item => item.Enabled && item.Ready && !item.Draining).length} enabled</p></div>
      <input className="plugin-search" aria-label={text.search} placeholder={text.search} value={pluginFilter} onChange={event => setPluginFilter(event.target.value)} />
    </div>
    <div className="table-wrap plugin-catalog-table">
      <table aria-label={text.registered}>
        <thead><tr><th>Plugin</th><th>Version</th><th>Status</th><th>Capabilities</th><th>Actions</th></tr></thead>
        <tbody>{rows.map(manifest => {
          const key = manifest.ID + "@" + manifest.Version;
          const installation = installed.get(key);
          const enabled = installation?.Enabled && installation.Ready && !installation.Draining;
          const status = !installation ? "Not installed" : installation.Draining ? "Draining" : enabled ? text.ready : installation.Ready ? text.disabled : text.notReady;
          return <tr key={key}>
            <td><Link className="table-title" href={projectUIURL(projectID, "plugins", key)}>{manifest.Title || manifest.ID}</Link><small className="muted table-subtitle">{manifest.ID}</small><small className="table-description">{manifest.Description}</small></td>
            <td><code>{manifest.Version}</code></td>
            <td><span className={"status-pill " + (enabled ? "is-success" : installation?.Draining ? "is-warning" : "is-inactive")}>{status}</span></td>
            <td>{manifest.Capabilities?.length ?? 0}</td>
            <td><div className="table-actions">
              {!installation ? <button type="button" onClick={() => setPending({ manifest, projectID })}>{text.install}</button> :
                !installation.Ready ? <button type="button" onClick={() => runPluginAction("plugin.readiness", installation)}>{text.markReady}</button> :
                !installation.Enabled && !installation.Draining ? <button type="button" onClick={() => runPluginAction("plugin.enable", installation)}>{text.enable}</button> : null}
              <Link className="table-action-link" href={projectUIURL(projectID, "plugins", key)} aria-label={"View " + manifest.ID}>Manage</Link>
            </div></td>
          </tr>;
        })}</tbody>
      </table>
      {!rows.length && <p className="muted" role="status">{query ? "No matching plugins." : text.empty}</p>}
    </div>
  </section>;
}
