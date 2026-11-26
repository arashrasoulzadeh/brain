import type { SystemManifest } from "../lib/admin-types";

type PluginConsoleProps = {
  visible: boolean;
  title: string;
  description: string;
  operation: string;
  onOperationChange: (value: string) => void;
  operations: readonly string[];
  payload: string;
  onPayloadChange: (value: string) => void;
  payloadLabel: string;
  sendLabel: string;
  onRun: () => void;
  message: string;
  manifests?: SystemManifest[];
};

export function PluginConsole({
  visible, title, description, operation, onOperationChange, operations, payload, onPayloadChange,
  payloadLabel, sendLabel, onRun, message, manifests = [],
}: PluginConsoleProps) {
  return (
    <section className={`section ${visible ? "" : "page-hidden"}`}>
      <div className="section-heading">
        <div>
          <p className="eyebrow">{title}</p>
          <h2>{title}</h2>
          <p className="muted">{description}</p>
        </div>
      </div>
      <div className="playground">
        <label>
          Operation
          <select value={operation} onChange={(event) => onOperationChange(event.target.value)}>
            {operations.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
        </label>
        <label>
          {payloadLabel}
          <textarea value={payload} onChange={(event) => onPayloadChange(event.target.value)} rows={10} />
        </label>
        <div className="playground-actions">
          <button type="button" onClick={onRun}>{sendLabel}</button>
          <small>{message}</small>
        </div>
      </div>
      {manifests.length > 0 && (
        <div className="plugin-grid">
          {manifests.map((manifest) => (
            <article className="plugin" key={`${manifest.ID}@${manifest.Version}`}>
              <h3>{manifest.Title}</h3>
              <p className="muted">{manifest.Description}</p>
              <div className="route-list">
                {manifest.Routes?.map((route) => (
                  <div key={route.Operation}>
                    <code>{route.Operation}</code>
                    <small>{route.Capability}</small>
                  </div>
                ))}
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
