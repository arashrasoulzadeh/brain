# Dashboard

Host-only operator control plane. [Authoritative manifest](../../plugins/dashboard/plugin.yaml).

## Developer

Manifest routes marked host.admin are reserved for trusted adapters and must not be invoked as generic workers. Use the HTTP admin command contract and shared admin SDK rather than duplicating transport and authorization logic.

## Administrator

Sign in to the Next.js admin. Select an active project, review manifest metadata, install, mark ready, grant, enable and inspect audit. Preview dependency impact before disabling; destructive actions require confirmation.

## Example payloads

POST /admin/actions: {"operation":"project.create","project_id":"dev"}. For complete command fields see [HTTP API](../http-api.md).

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Dashboard manifest routes are not the complete HTTP command inventory; the manifest includes plugin.uninstall, while the actual HTTP action is plugin.remove. Use the mounted API contract. Do not expose the operator bearer or assume a worker policy grant makes host-only routes callable.
