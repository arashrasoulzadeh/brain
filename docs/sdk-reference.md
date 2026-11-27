# TypeScript SDK reference

The trusted shared SDK lives in [plugins/sdk-typescript/src/index.ts](../plugins/sdk-typescript/src/index.ts). Import declared runtime dependencies from this boundary (or explicitly permitted helper modules). A local helper or request builder never grants authority to a plugin.

| Area | Exported APIs | Contract |
| --- | --- | --- |
| Request construction | `request`, `pluginRequest`, `createResourceRequest`, `createManifestRequest` | Versioned request serialization; explicit resource identity overrides a conflicting serialized payload field. |
| Dispatch | `createRouteHandler`, `createHostHandler`, `createFunctionHandler`, `createOperationRoutes`, `createRequestDispatcher`, `createRouteDispatcher` | Dispatch declared handlers only. Trusted execution context is checked by host-facing handlers. |
| Context and scope | `validateProjectId`, `projectRoomID`, `requireProjectInput`, `requireOperator` | Validate project/context, reject mismatched identity and enforce required operator permission. |
| Input validation | `validateIdentifier`, `requiredString`, `trimmedString`, `validateBoundedInteger`, `requireAllowedValues` | Shared identifier, string, integer and allowed-value checks. |
| In-memory project maps | `createProjectMap`, `createProjectSet`, `createProjectLeaseStore` | Isolate local values by project and key; not durable storage. |
| Execution primitives | `commitWithGrant`, `createCommitRollback`, `createTransactionGate`, `runCappedRetries`, `runWithTimeout`, `createFixedWindowLimiter` | Local bounded execution and commit helpers; host remains final authority. |
| Declarative helpers | `validateModel`, `transitionState`, `callHelper`, `hasHelper`, `helperNames` | Exact field validation, state transition calculation and allowlisted pure helper evaluation. |
| PvP | `CommandDeduplicator`, `assertMatchMember`, `nextTurn`, `createRevisionRuntime`, `commitRevision`, `replayHistory`, `expireTurn`, `pauseTurn`, `resumeTurn`, `privatePlayerView`, `parseMatchState` | Reusable game-neutral turn and revision primitives, not an exposed network service. |

## Context types

`PluginContext` contains the trusted project projection, plugin id/version, execution operation/invocation ID, actor/caller/permissions, typed config, pure helpers, dependency snapshot and observation publishing API. The host supplies production context; optional members support legacy fixtures. Use [authentication context](plugin-auth-context.md) to understand actor vs caller.

`RpcRequest` / `WorkerParams` serialize the versioned JSON-RPC invocation. `RpcResponse` contains protocol, success, payload, optional observations or error code. `PluginRoute`, `PluginHandler`, `PluginFunction`, `ContextualPluginFunction`, `PayloadHandler`, `HostInvoker` and `HostExecutionContext` describe dispatch boundaries. `Observation<T>` and `ObservationIntent` distinguish trusted received events from emitted intents. `ProjectContext`, `InvocationActor`, `InvocationPermission` and `InvocationAuth` describe the trusted security projection. `DeclarativeResponse`, `DeclarativePipeline`, `ModelSchema`, `StateMachine`, `PureHelperName`, `TurnMatchState`, `RevisionRuntime` and `CommittedCommand` describe bounded data contracts.

## Pure helper names

`string.trim`, `string.nonempty`, `string.lower`, `string.upper`, `number.integer`, `boolean.require`, `value.identity`. Unknown helper names or invalid argument types are rejected. The same allowlist is shared by declarative pipelines and the function VM.

Use `helper: sdk.context` for new project-bound functions. See [developer guide](developers.md) for manifest and event examples and [declarative runtime](declarative-runtime.md) for source-free plugins. Direct SDK tests are not evidence of host-level authorization, persistence or network wiring.

`ManifestOperations` and `ManifestOperation` are generated from the discovered plugin manifests. Regenerate after editing route declarations with `go run ./cmd/manifest-sdk`. `createManifestRequest` binds the plugin ID and preserves explicit resource identity; it does not confer authority.
