export type Project = { ID: string; Active: boolean };
export type PluginManifest = {
  ID: string;
  Title: string;
  Description: string;
  Version: string;
  Capabilities: string[];
  ConfigSchema?: Record<string, string>;
  Admin?: AdminSpec | null;
};
export type AdminField = { name: string; type: "string" | "number" | "boolean" | "json"; required?: boolean };
export type AdminResultSpec = { type: "table" | "detail" | "metric" | "chart" | "json" | "code" | "logs" | "grid" | "tabs"; title?: string; path?: string; columns?: string[]; page_size?: number; label_field?: string; value_field?: string; children?: AdminResultSpec[] };
export type AdminAction = { title: string; operation: string; fields?: AdminField[]; read_only?: boolean; result?: AdminResultSpec };
export type AdminInventorySpec = { kind: "state_keys" | "file_keys" | "table_names"; capability: string; title: string; select_field?: string };
export type AdminInventoryRow = { key: string; size_bytes?: number; published?: boolean };
export type AdminInventoryPage = { rows: AdminInventoryRow[]; next?: string };
export type AdminPage = { id: string; title: string; actions: AdminAction[]; extends?: string; refresh_on?: string[]; legacy_route?: string; inventory?: AdminInventorySpec; extends_slot?: "link" | "tab"; extension_slots?: "tab"[] };
export type AdminSpec = { version: number; pages: AdminPage[] };
export type PluginStatus = {
  ProjectID: string;
  PluginID: string;
  Version: string;
  Enabled: boolean;
  Ready: boolean;
  WorkerReady?: boolean;
  Draining: boolean;
  Grants?: string[];
};
export type AuditRecord = {
  source: string;
  action: string;
  actor_id?: string;
  project_id?: string;
  plugin_id?: string;
  version?: string;
  success: boolean;
  detail?: string;
};
export type FunctionDraft = { project_id: string; name: string; source: string; created_at: string };
export type SystemManifest = PluginManifest & {
  Routes: { Operation: string; Capability: string }[];
  ExtensionPoints: string[];
  SourceFiles: string[];
  Imports: string[];
  PluginImports: string[];
  Dependencies: string[];
};
export type Language = "en" | "fa";
