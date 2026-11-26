import type { AdminCopy } from "../lib/admin-copy";
import type { SystemManifest } from "../lib/admin-types";

export function SystemPanel({ activePage, text, system }: {
  activePage: string;
  text: AdminCopy;
  system: SystemManifest[];
}) {
  return (
    <section className={`section ${activePage !== "system" ? "page-hidden" : ""}`} id="system">
      <div className="section-heading">
        <div>
          <p className="eyebrow">{text.system}</p>
          <h2>{text.system}</h2>
        </div>
        <span className="count">{system.length}</span>
      </div>
      <div className="plugin-grid">
        {system.map((manifest) => (
          <article className="plugin" key={`system-${manifest.ID}@${manifest.Version}`}>
            <h3>{manifest.Title || manifest.ID}</h3>
            <p className="muted">{manifest.Description}</p>
            <small>
              {manifest.Routes?.length ?? 0} {text.routes} ·{" "}
              {manifest.ExtensionPoints?.length ?? 0} {text.extensions} ·{" "}
              {manifest.SourceFiles?.length ?? 0} {text.files}
            </small>
            <div className="capabilities">
              {manifest.Capabilities?.map((capability) => (
                <span key={capability}>{capability}</span>
              ))}
            </div>
            <div className="route-list">
              {manifest.Routes?.map((route) => (
                <div key={route.Operation}>
                  <code>{route.Operation}</code>
                  <small>{route.Capability}</small>
                </div>
              ))}
            </div>
            <details>
              <summary>
                {text.files} / {text.imports} / {text.dependencies}
              </summary>
              <small>{manifest.SourceFiles?.join(" · ")}</small>
              <small>{manifest.Imports?.join(" · ")}</small>
              <small>
                {[...(manifest.Dependencies ?? []), ...(manifest.PluginImports ?? [])].join(
                  " · ",
                ) || "—"}
              </small>
            </details>
          </article>
        ))}
      </div>
    </section>
  );
}
