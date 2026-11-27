# Functions

## Developer

Use the [function authoring screens](../admin-pages/functions.md) to create and publish a version. The manifest route `functions.call` takes a name, optional version and input; this route's public permission validation is not a substitute for authorization on the actual published-function endpoint. The host checks publication, method and session or policy before executing the source. Never execute arbitrary user source outside the supervised runtime.

## Administrator

Separate authoring/publishing grants from invocation, inspection and logs. Review declared project and actor permissions for every published version. Validate the resource name, test the exact version, inspect sanitized logs, and roll back using the documented lifecycle. Review worker isolation and timeout evidence before exposing a public endpoint.
