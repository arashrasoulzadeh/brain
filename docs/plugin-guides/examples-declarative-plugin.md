# Declarative catalog example

## Developer

This example has no executable source files. `catalog.create`, `catalog.list` and `catalog.publish` use the host model; `catalog.project` maps the trusted project ID through `sdk.response`. The item model contains a string name and integer count, initially draft, with a publish transition. See [declarative runtime](../declarative-runtime.md).

## Administrator

This example is not automatically installed from the examples directory. Use a fixture environment, register the manifest through the supported discovery path, grant catalog read/write, and verify rejected invalid fields, stale revisions and unauthorized transitions. Do not add executable code to a declarative-only manifest.
