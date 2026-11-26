# Plugin project context and observation migration

Status: in progress. Do not mark an item complete until its focused and full CI gates pass.

## Runtime
- [ ] Verify shared worker injection for sdk.dispatch, sdk.request, sdk.context, sdk.response, handler and mapped exports; remove remaining legacy payload-only production exports.
- [ ] Resolve the existing Project model from trusted host state (including active status); expose a read-only safe projection, never caller-provided identity.
- [ ] Require context for every project-bound production function and propagate through functions, workflow, scheduled jobs and external workers. Pure helpers and request serializers are not authorization boundaries.
- [ ] Add host-validated manifest event and observation declarations, explicit subscriber grants and event payload contracts.
- [ ] Add SDK plugins.observe.fire(event, payload) with implicit publisher and project; stage bounded JSON events, dispatch only after successful host commit, never from failed invocations.
- [ ] Resolve subscribers from enabled installations in the same project, recheck readiness, permissions and lifecycle, invoke through shared registry; isolate failure, bound recursion and fanout, audit without sensitive payloads.
- [ ] Support YAML observers via existing manifest function dispatch, without importing publisher code. Define delivery and retry semantics.

## All plugin migration inventory
Each row requires context and relevant sanitized domain events, focused tests and manifest updates. No artificial events for pure reads; do not emit secrets or raw authentication payloads.
- [ ] analytics — capture context, subscribe to rooms.connected; avoid observation loops.
- [ ] arena — match started, finished; verify actual route exports.
- [ ] auth — account created, session established; exclude credentials, OTPs, tokens.
- [ ] dashboard — committed lifecycle changes; separate privileged system operations from project functions; verify actual route exports.
- [ ] files — uploaded, published; exclude file contents.
- [ ] functions — published, invoked; include scheduled and workflow paths, not source or arbitrary inputs.
- [ ] invites — created, accepted, cancelled; exclude invite codes; verify actual route exports.
- [ ] matchmaking — queued, assigned, cancelled.
- [ ] rooms — created, connected, reconnected, result committed; actual membership/connection semantics.
- [ ] subscriptions — opened, closed; exclude ticket secrets and private content; verify actual route exports.
- [ ] tables — record written; avoid raw record values unless explicitly authorized.
- [ ] template — context-first example and observer example.

## Exit gates
- [ ] Two-project isolation, forged identity, failed commit, disabled/revoked subscriber, recursive events, bounded payload and fanout, manifest validation, no observer dependency on publisher.
- [ ] Run npm run lint:ts, npm run typecheck, npm run test:ts, go test ./... and sandboxed worker tests.
- [ ] Update developer/operator documentation and changelog; no completion claim until verification.

## Implementation progress (pending full verification)

- Shared worker context, trusted project resolver, manifest contracts, and post-commit observer dispatch implemented on this branch.
- Auth, files, matchmaking, rooms, tables, and analytics have observation integration; template has a context-first publisher example.
- Functions draft execution now receives a read-only project context as the second argument. The plugin call route emits an acceptance event only; actual function publication, execution, workflow and scheduler events remain open.
- Arena, dashboard, invites and subscriptions have existing manifest routes without matching executable exports; implement their actual domain adapters before declaring migration complete.
- Analytics observation capture remains in-memory; durable storage and delivery/retry policies are still open. Initial delivery is synchronous, bounded, best-effort, and non-recursive.

- Arena: context-first start/input/snapshot/result adapters and sanitized start/finish events added to its existing state machine; cross-project and invalid-transition regression coverage added. Operator route only; realtime actor binding, durable results, and CI verification remain open.

- Invites: operator route adapters added with host project context, random identifiers, bounded TTL, replay/cancellation checks, and sanitized lifecycle observations. Client actor authorization and durable invitation storage are still open.

- Dashboard: declared host-only admin routes (never executable in workers); authenticated project lifecycle actions publish sanitized host observations after success. Global readiness and inactive-project actions are not delivered to project workers. Host-origin schemas prevent worker spoofing.

- Subscriptions: operator-only open/receive/resync/close adapters for the real `rooms.connected` observation stream, with bounded retention, project/channel-bound polling tickets and sanitized open/close events. Existing WebSocket actor authorization and record-change channel integration remain separate unfinished work.

- Functions: successful host publication and actual VM invocation now emit host-origin observations (not the declarative acknowledgment routes). VM receives read-only project, plugin identity and execution metadata. Scheduled/workflow execution is not yet a production runner; the existing helpers remain to be integrated.

- Mandatory context: production worker descriptors require the host project model, direct legacy exports receive context as a second argument, root forged project identity is denied, and the credential-based worker boundary re-resolves project state. The external protobuf contract now carries project snapshots and observation intents; external transport implementation is still not wired.

- Durable delivery: SQLite-backed bounded per-project queue, asynchronous host drain, stable event IDs, three attempts with backoff, reauthorization before delivery, and optional durable observer receipts. Analytics opts in to receipt persistence. Publisher mutation and queue enqueue are currently separate commits; a host crash between them can lose an event. Multi-host leasing and exactly-once side effects are not guaranteed.

- SQLite-backed plugin mutations now commit their prepared storage effect and observation delivery inserts in one transaction. A failed outbox insert rolls back that durable mutation. Host-side admin/functions and in-memory domain effects still use post-commit enqueue; no transactional guarantee is claimed for those paths.
