export type SchemaChange = { kind: "add_column"; name: string; type: "string" | "number" | "boolean" | "date" } | { kind: "create_index"; name: string; columns: readonly string[] };
export type SchemaChangePlan = { projectId: string; table: string; change: SchemaChange };
export type SchemaPublication = SchemaChangePlan & { actorId: string; publishedAt: number };
