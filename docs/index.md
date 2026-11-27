# Brain documentation

<div class="brain-hero" markdown="1">

<span class="brain-kicker">PROJECT-ISOLATED BACKEND · PLUGIN PLATFORM</span>

## Build apps, websites, and games on a composable backend.

Brain combines a Go host with versioned TypeScript plugins, scoped storage, and a single operator control plane. Build a feature as a plugin; let the host enforce project identity, lifecycle, and access.

[Get started](developers.md){ .md-button .brain-primary } [Operate Brain](admins.md){ .md-button .brain-secondary }

</div>

## Architecture at a glance

<div class="grid cards" markdown="1">

- **Trusted host**

    Go owns transport, identity, authorization, project storage, and lifecycle.

    [HTTP API and host routes](http-api.md)

- **Versioned plugins**

    TypeScript and declarative plugins expose explicit, manifest-defined operations and observation hooks.

    [Plugin communication](plugin-communication.md)

- **One control plane**

    The admin UI exposes only the installed, ready, and authorized capabilities of each project.

    [Administration guide](admins.md)

- **Portable clients**

    Browser, Flutter and Kotlin clients use the same project-scoped protocol.

    [Browser quickstart](browser-quickstart.md)

</div>

## System flow

```mermaid
flowchart LR
  C[Client] --> T[Go transport]
  T --> P[Host-authenticated project context]
  P --> B[Permission broker]
  B --> W[TypeScript plugin worker]
  W --> S[Scoped storage / service]
  B --> A[Audit and lifecycle state]
  W --> R[Scoped response]
```

The plugin never supplies the authoritative project identity. The host binds it before authorization and rechecks lifecycle, grants, versions, and single-use invocation credentials at commit time.

## Start with a clear boundary

!!! tip "Local development"
    Do not run Docker during ordinary local development. Use in-process tests and SQLite-compatible local checks. PostgreSQL and Compose checks are explicit integration or release work.

- [Deployment runbook](deployment-runbook.md): release checks and rollback evidence.
- [Restore runbook](restore-runbook.md): recovery procedures.
- [Phase options](phase-options.md): decisions and alternatives.
- [Quality and coverage](quality.md): reproducible checks and coverage targets.
