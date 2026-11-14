import { createInvocationCommitState } from "./invocation-commit-state.js";
import { InvocationCommitWriter } from "./invocation-commit-writer.js";
import { InvocationCommitReader } from "./invocation-commit-reader.js";
import { InvocationResultReader } from "./invocation-result-reader.js";
import { InvocationExistenceReader } from "./invocation-existence-reader.js";
export function createInvocationCommitStore() { const state = createInvocationCommitState(); const writer = new InvocationCommitWriter(state); const effects = new InvocationCommitReader(state); const result = new InvocationResultReader(state); const existence = new InvocationExistenceReader(state); return { commit: writer.commit.bind(writer), effects: effects.effects.bind(effects), result: result.result.bind(result), has: existence.has.bind(existence), setLifecycleRevision: (revision: number) => { state.lifecycleRevision = revision; } }; }
