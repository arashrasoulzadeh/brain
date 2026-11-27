import { PluginManageLink } from "./plugin-manage-link";
import { installationHealth } from "../lib/admin-observability";
import type { AdminCopy } from "../lib/admin-copy";
import type { PluginManifest, PluginStatus } from "../lib/admin-types";

export function MonitoringPanel({ activePage, text, manifests, installations }: {
  activePage: string;
  text: AdminCopy;
  manifests: PluginManifest[];
  installations: PluginStatus[];
}) {
  return (
    <section className={`section ${activePage !== "monitoring" ? "page-hidden" : ""}`}>
      <div className="section-heading">
        <div>
          <p className="eyebrow">{text.monitoring}</p>
          <h2>{text.monitoring}</h2>
        </div>
        <span className="count">{manifests.length}</span>
      </div>
      <p className="muted">{installations.length} installed · {installations.filter(item => { const manifest = manifests.find(value => value.ID === item.PluginID && value.Version === item.Version); return manifest && installationHealth(manifest, item) === "ready"; }).length} healthy</p>
      {!manifests.length && <p className="muted">No registered plugins.</p>}
      <div className="audit-list">
        {manifests.map((manifest) => {
          const installation = installations.find(
            (candidate) =>
              candidate.PluginID === manifest.ID && candidate.Version === manifest.Version,
          );
          const health = installationHealth(manifest, installation);
          const state = { absent: "not installed", draining: "draining", pending: text.notReady, disabled: text.disabled, grants: "missing capability grants", worker: "worker unavailable", ready: text.ready }[health];
          const healthy = health === "ready";
          const statusClass = !installation ? "neutral" : healthy ? "success" : "failure";
          return (
            <div className="audit" key={`monitor-${manifest.ID}@${manifest.Version}`}>
              <span className={statusClass}>{healthy ? "✓" : !installation ? "—" : "!"}</span>
              <div>
                <strong>{manifest.Title || manifest.ID}</strong>
                <small>
                  {manifest.ID}@{manifest.Version} · {state}
                </small>
              </div>
              {installation && <PluginManageLink projectID={installation.ProjectID} pluginID={installation.PluginID} version={installation.Version} />}
            </div>
          );
        })}
      </div>
    </section>
  );
}
