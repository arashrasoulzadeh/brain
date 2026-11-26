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
      <div className="audit-list">
        {manifests.map((manifest) => {
          const installation = installations.find(
            (candidate) =>
              candidate.PluginID === manifest.ID && candidate.Version === manifest.Version,
          );
          const state = !installation
            ? "not installed"
            : installation.Draining
              ? "draining"
              : !installation.Ready
                ? text.notReady
                : !installation.Enabled
                  ? text.disabled
                  : installation.WorkerReady === false
                    ? "worker unavailable"
                    : text.ready;
          const healthy = Boolean(
            installation?.Enabled &&
            installation?.Ready &&
            installation.WorkerReady !== false &&
            !installation.Draining,
          );
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
            </div>
          );
        })}
      </div>
    </section>
  );
}
