export type Project = { ID: string; Active: boolean; Name?: string; Slug?: string; Settings?: Record<string, unknown> };
export type PluginManifest = {
  ID: string;
  Title: string;
  Description: string;
  Version: string;
  Capabilities: string[];
  Dependencies?: string[];
  ConfigSchema?: Record<string, string>;
  Admin?: AdminSpec | null;
};
export type AdminField = { name: string; type: "string" | "number" | "boolean" | "json"; required?: boolean };
export type AdminResultSpec = { type: "table" | "detail" | "metric" | "chart" | "json" | "code" | "logs" | "grid" | "tabs"; title?: string; path?: string; columns?: string[]; page_size?: number; label_field?: string; value_field?: string; children?: AdminResultSpec[] };
export type DeclaredModel = { version: number; fields: Record<string, "string" | "integer" | "boolean">; initial?: string; transitions?: Record<string, { from: string; to: string }> };
export type AdminModelAction = "create" | "read" | "list" | "update" | "delete" | "transition";
export type AdminAction = { title: string; operation: string; fields?: AdminField[]; read_only?: boolean; result?: AdminResultSpec; model_action?: AdminModelAction; model_transition?: string };
export type AdminInventorySpec = { kind: "state_keys" | "file_keys" | "table_names"; capability: string; title: string; select_field?: string };
export type AdminInventoryRow = { key: string; size_bytes?: number; stored_bytes?: number; published?: boolean; filename?: string; extension?: string; content_type?: string; updated_at?: string };
export type AdminInventoryPage = { rows: AdminInventoryRow[]; next?: string };
export type AdminPage = { id: string; title: string; actions: AdminAction[]; extends?: string; refresh_on?: string[]; legacy_route?: string; inventory?: AdminInventorySpec; extends_slot?: "link" | "tab"; extension_slots?: "tab"[]; model?: { name: string; schema: DeclaredModel } };
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
  ConfiguredFields?: string[];
  PublicConfig?: Record<string, string | number | boolean>;
  MaxUploadBytes?: number;
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

export type ProjectHTTPRoute = { plugin_id: string; version: string; method: string; path: string; function: string; capability: string; operation: string; auth: "public" | "user" | "operator" };
