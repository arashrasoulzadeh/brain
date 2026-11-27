# Brain documentation

<div class="brain-hero" markdown="1">

**PROJECT-ISOLATED BACKEND · PLUGIN PLATFORM**

## Build apps, websites and games on a composable backend.

Brain combines a Go host, versioned TypeScript plugins, scoped durable storage and one operator control plane.

[Start developing](developers.md){ .md-button .md-button--primary } [Operate Brain](admins.md){ .md-button }

</div>

## Start here

<div class="grid cards" markdown="1">

- **Developer quickstart**

    [Set up locally](developers.md), [write plugins](sdk-reference.md), and [build functions](functions-runtime.md).

- **HTTP and security**

    [API reference](http-api.md), [auth context](plugin-auth-context.md), and [plugin communication](plugin-communication.md).

- **Operate the host**

    [Runtime settings](runtime-operations.md), [admin pages](admin-pages/README.md), and [deployment](deployment-runbook.md).

- **Integrate clients**

    [JavaScript](browser-quickstart.md), [Flutter](flutter-quickstart.md), and [Kotlin](../clients/android/README.md).

</div>

## Architecture

```mermaid
flowchart LR
  C[Client or operator] --> H[Go transport]
  H --> P[Project and authenticated actor]
  P --> B[Permission broker]
  B --> W[Declared plugin worker]
  W --> S[Scoped storage commit]
  S --> O[Observation outbox]
  O --> A[Authorized subscriber]
```

The host binds project identity before authorization and rechecks grants, lifecycle, versions and invocation credentials at commit. A plugin cannot authorize itself.

## Current scope

The running application uses SQLite and sandboxed TypeScript workers, with a separate restricted user-function VM. Generic end-user HTTP RPC, distributed observation delivery, PostgreSQL application storage and full production Go/WASM function deployment remain separate integration/release work. The [roadmap](../roadmap.md) tracks these items; API and plugin references describe what is implemented now.

See [quality and coverage](quality.md) for verification and [restore procedures](restore-runbook.md) for recovery.
