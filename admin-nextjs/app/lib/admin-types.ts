export type Project = { ID: string; Active: boolean };
export type PluginManifest = {
  ID: string;
  Title: string;
  Description: string;
  Version: string;
  Capabilities: string[];
  ConfigSchema?: Record<string, string>;
};
export type PluginStatus = {
  ProjectID: string;
  PluginID: string;
  Version: string;
  Enabled: boolean;
  Ready: boolean;
  WorkerReady?: boolean;
  Draining: boolean;
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
