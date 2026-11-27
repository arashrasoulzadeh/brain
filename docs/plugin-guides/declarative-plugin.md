# Declarative catalog example

Source-free manifest. [Authoritative manifest](../../examples/declarative-plugin/plugin.yaml).

## Developer

Define bounded model fields and transitions, pure sdk.response or sdk.pipeline routes and fixed host.call dependencies. The catalog example declares item name and count with draft-to-published transition. No TypeScript entrypoint is required.

## Administrator

This directory is outside the normal discovery root. To test, copy into a disposable plugin root and restart discovery, then install and grant catalog.read and catalog.write. Keep a current record revision before transitions.

## Example payloads

catalog.create: {"id":"one","fields":{"name":"Example","count":1}}; catalog.list: {}; catalog.publish: {"id":"one","revision":"<current-revision>"}; catalog.project: {}.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Declarative helpers cannot perform arbitrary effects, authorization, imports or nested general dependency calls.
