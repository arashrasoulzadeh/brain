import type { AdminInventoryRow, AuditRecord, PluginManifest, PluginStatus } from "./admin-types.js";

export type AuditOutcome = "all" | "success" | "failure";
export function filterAudit(records: readonly AuditRecord[], search: string, outcome: AuditOutcome, source: string): AuditRecord[] {
  const needle = search.trim().toLowerCase();
  return records.slice().reverse().filter(record =>
    (outcome === "all" || record.success === (outcome === "success")) &&
    (!source || record.source === source) &&
    [record.action, record.source, record.project_id, record.plugin_id, record.version, record.actor_id, record.detail, record.occurred_at]
      .some(value => value?.toLowerCase().includes(needle)),
  ).sort((a, b) => (Date.parse(b.occurred_at ?? "") || 0) - (Date.parse(a.occurred_at ?? "") || 0));
}

export function filterInventory(rows: readonly AdminInventoryRow[], search: string): AdminInventoryRow[] {
  const needle = search.trim().toLowerCase();
  return rows.filter(row => [row.key, row.filename, row.extension, row.content_type].some(value => value?.toLowerCase().includes(needle)));
}

export function installationHealth(manifest: PluginManifest, installation?: PluginStatus): "absent" | "draining" | "pending" | "disabled" | "grants" | "worker" | "ready" {
  if (!installation) return "absent";
  if (installation.Draining) return "draining";
  if (!installation.Ready) return "pending";
  if (!installation.Enabled) return "disabled";
  if ((manifest.Capabilities ?? []).some(capability => !installation.Grants?.includes(capability))) return "grants";
  return installation.WorkerReady === true ? "ready" : "worker";
}
