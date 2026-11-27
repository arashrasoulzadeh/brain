import type { AdminCopy } from "../lib/admin-copy";
import type { AuditRecord } from "../lib/admin-types";

export function AuditPanel({ activePage, text, audit, auditFilter, setAuditFilter }: {
  activePage: string;
  text: AdminCopy;
  audit: AuditRecord[];
  auditFilter: string;
  setAuditFilter: (value: string) => void;
}) {
  return (
    <section className={`section ${activePage !== "audit" ? "page-hidden" : ""}`} id="audit">
      <div className="section-heading">
        <div>
          <p className="eyebrow">{text.audit}</p>
          <h2>{text.audit}</h2>
        </div>
        <span className="count">{audit.length}</span>
      </div>
      <input
        className="plugin-search"
        aria-label={text.searchAudit}
        placeholder={text.searchAudit}
        value={auditFilter}
        onChange={(event) => setAuditFilter(event.target.value)}
      />
      <div className="audit-list">
        {audit.length ? (
          audit
            .slice()
            .reverse()
            .filter((record) =>
              `${record.action} ${record.source} ${record.project_id ?? ""} ${record.plugin_id ?? ""} ${record.actor_id ?? ""} ${record.detail ?? ""}`
                .toLowerCase()
                .includes(auditFilter.toLowerCase()),
            )
            .map((record, index) => (
              <div className="audit" key={`${record.action}-${index}`}>
                <span className={record.success ? "success" : "failure"}>
                  {record.success ? "✓" : "!"}
                </span>
                <div>
                  <strong>{record.action}</strong>
                  <small>
                    {record.source}
                    {record.plugin_id ? ` · ${record.plugin_id}@${record.version}` : ""}
                    {record.project_id ? ` · ${record.project_id}` : ""}
                    {record.actor_id ? ` · ${record.actor_id}` : ""}
                    {record.success ? " · Success" : " · Failed"}
                  </small>
                  {record.detail && <p className="muted">{record.detail}</p>}
                </div>
              </div>
            ))
        ) : (
          <p className="muted">{text.noAudit}</p>
        )}
      </div>
    </section>
  );
}
