# Plugin project context and observation migration

Status: implementation in PR #6; final integration gates pending. Checkboxes
are reserved for verified work, not merely committed code.

## Runtime

- [ ] Verify all worker dispatch modes, the trusted active project projection,
  and forged project identity rejection. Production descriptors require context.
- [ ] Verify manifest event schemas, subscriber grants, implicit publisher
  identity, bounded fanout/payloads, failure isolation and recursion prevention.
- [ ] Verify durable SQLite queue, stable event IDs, bounded retries, delivery
  reauthorization and project-scoped optional receipts.
- [ ] Verify atomic durable plugin effect + outbox insertion, rollback on failure,
  and credential-bound external worker registry integration.

## Plugin inventory

All entries have context/event or observer code implemented; final verification
is not yet complete.

| Plugin | Implemented behavior |
| --- | --- |
| Analytics | Context-bound direct capture; captured event; persisted rooms.connected observer |
| Arena | Real start/input/snapshot/result adapters; match start and finish |
| Auth | Account/session actions; sanitized events without credentials |
| Dashboard | Host-only authenticated lifecycle routes and committed host events |
| Files | Upload/publish events; authoritative storage state |
| Functions | Actual host publication/invocation events; context in function VM; project-bound scheduling/workflow helper facades |
| Invites | Create/accept/cancel; random IDs and bounded TTL; sanitized events |
| Matchmaking | Queue/assign/cancel; authoritative ticket state |
| Rooms | Create/connect/reconnect/result; authoritative room state |
| Subscriptions | Actual rooms.connected observer; operator polling open/receive/resync/close |
| Tables | Committed record writes; no raw record payload in observations |
| Template | Publisher and subscriber examples without cross-plugin imports |

## Exit gates

- [ ] Final Go, TypeScript lint/typecheck/tests, sandboxed worker, admin and
  production-runtime CI at the latest branch head.
- [ ] Actual packaged plugin startup and Rooms-to-Analytics/Subscriptions
  delivery smoke test.
- [ ] Final review and merge only after required checks.

Developer, operator and plugin communication documentation has been updated.
The changelog is generated automatically from commits on main.

## Remaining production gaps (not marked complete)

- Existing scheduled job/workflow orchestration helpers are now project bound,
  but the production scheduler/executor itself is not wired.
- External worker credential calls and protobuf schema use the shared registry;
  the external gRPC transport itself is not wired.
- End-user realtime actor/membership authorization and record-change channels
  are separate from the current operator polling/arena/invitation adapters.
- In-memory domain state (including arena and invitations) is not durable.
- SQLite-backed plugin effects and their observation outbox insert atomically.
  Host admin/function and in-memory domain effects still enqueue separately.
- Single-host queue, bounded retries; possible duplicate delivery and no
  distributed leasing or exactly-once side effects. Optional Analytics receipts
  persist, but receipt retention/export management is separate work.
