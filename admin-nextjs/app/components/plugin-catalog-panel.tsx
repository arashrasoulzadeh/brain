import type { AdminCopy } from "../lib/admin-copy";
import type { PluginManifest, PluginStatus } from "../lib/admin-types";

export function PluginCatalogPanel({ activePage, text, manifests, installations, pluginFilter, setPluginFilter, installPlugin, runPluginAction }: {
  activePage: string;
  text: AdminCopy;
  manifests: PluginManifest[];
  installations: PluginStatus[];
  pluginFilter: string;
  setPluginFilter: (value: string) => void;
  installPlugin: (manifest: PluginManifest) => void;
  runPluginAction: (operation: string, installation: PluginStatus) => void;
}) {
  return (
    <>
      <section className="section page-hidden" id="projects">
        <div className="section-heading">
          <div>
            <p className="eyebrow">{text.plugins}</p>
            <h2>{text.registered}</h2>
          </div>
          <span className="count">{manifests.length}</span>
        </div>
        <input
          className="plugin-search"
          placeholder={text.search}
          value={pluginFilter}
          onChange={(event) => setPluginFilter(event.target.value)}
        />
        <div className="plugin-grid">
          {manifests.filter((manifest) =>
            `${manifest.ID} ${manifest.Title} ${manifest.Description}`
              .toLowerCase()
              .includes(pluginFilter.toLowerCase()),
          ).length ? (
            manifests
              .filter((manifest) =>
                `${manifest.ID} ${manifest.Title} ${manifest.Description}`
                  .toLowerCase()
                  .includes(pluginFilter.toLowerCase()),
              )
              .map((manifest) => {
                const installation = installations.find(
                  (item) => item.PluginID === manifest.ID && item.Version === manifest.Version,
                );
                const installed = Boolean(installation);
                return (
                  <article className="plugin" key={`${manifest.ID}@${manifest.Version}`}>
                    <div>
                      <h3>
                        <a
                          href={`/plugins/${encodeURIComponent(`${manifest.ID}@${manifest.Version}`)}`}
                        >
                          {manifest.Title || manifest.ID}
                        </a>
                      </h3>
                      <p className="muted">{manifest.Description}</p>
                    </div>
                    <small>
                      {manifest.ID}@{manifest.Version}
                    </small>
                    <div className="capabilities">
                      {manifest.Capabilities?.map((capability) => (
                        <span key={capability}>{capability}</span>
                      ))}
                    </div>
                    {!installed && (
                      <button className="plugin-action" onClick={() => installPlugin(manifest)}>
                        {text.install}
                      </button>
                    )}
                    {installation && (
                      <div className="lifecycle-actions">
                        {!installation.Ready && (
                          <button onClick={() => runPluginAction("plugin.readiness", installation)}>
                            {text.markReady}
                          </button>
                        )}
                        {!installation.Enabled && installation.Ready && (
                          <button onClick={() => runPluginAction("plugin.enable", installation)}>
                            {text.enable}
                          </button>
                        )}
                        {installation.Ready && <small>{text.ready}</small>}
                      </div>
                    )}
                  </article>
                );
              })
          ) : (
            <p className="muted">{text.empty}</p>
          )}
        </div>
      </section>
      <section
        className={`section ${activePage !== "plugins" ? "page-hidden" : ""}`}
        id="installed-plugin-table"
      >
        <div className="section-heading">
          <div>
            <p className="eyebrow">{text.plugins}</p>
            <h2>{text.installed}</h2>
          </div>
          <span className="count">{installations.length}</span>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Plugin</th>
                <th>Version</th>
                <th>Status</th>
                <th>View</th>
              </tr>
            </thead>
            <tbody>
              {installations.map((installation) => (
                <tr key={`${installation.PluginID}@${installation.Version}`}>
                  <td>{installation.PluginID}</td>
                  <td>{installation.Version}</td>
                  <td>
                    {installation.Enabled && installation.Ready
                      ? text.ready
                      : installation.Ready
                        ? text.disabled
                        : text.notReady}
                  </td>
                  <td>
                    <a
                      className="view-icon"
                      aria-label={`View ${installation.PluginID}`}
                      title="View plugin"
                      href={`/plugins/${encodeURIComponent(`${installation.PluginID}@${installation.Version}`)}`}
                    >
                      ↗
                    </a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
      <section className={`section ${activePage !== "plugins" ? "page-hidden" : ""}`}>
        <div className="section-heading">
          <div>
            <p className="eyebrow">{text.plugins}</p>
            <h2>{text.registered}</h2>
          </div>
          <span className="count">{manifests.length}</span>
        </div>
        <input
          className="plugin-search"
          placeholder={text.search}
          value={pluginFilter}
          onChange={(event) => setPluginFilter(event.target.value)}
        />
        <div className="plugin-grid">
          {manifests
            .filter((manifest) =>
              `${manifest.ID} ${manifest.Title} ${manifest.Description}`
                .toLowerCase()
                .includes(pluginFilter.toLowerCase()),
            )
            .map((manifest) => {
              const installed = installations.some(
                (item) => item.PluginID === manifest.ID && item.Version === manifest.Version,
              );
              return (
                <article className="plugin" key={`catalog-${manifest.ID}@${manifest.Version}`}>
                  <div>
                    <h3>{manifest.Title || manifest.ID}</h3>
                    <p className="muted">{manifest.Description}</p>
                  </div>
                  <small>
                    {manifest.ID}@{manifest.Version}
                  </small>
                  <div className="capabilities">
                    {manifest.Capabilities?.map((capability) => (
                      <span key={capability}>{capability}</span>
                    ))}
                  </div>
                  <button
                    className="plugin-action"
                    disabled={installed}
                    onClick={() => installPlugin(manifest)}
                  >
                    {installed ? text.installedLabel : text.install}
                  </button>
                </article>
              );
            })}
        </div>
      </section>
    </>
  );
}
