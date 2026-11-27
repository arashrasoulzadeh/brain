# Create function

Route: `/{project}/ui/functions/new`

Choose a JavaScript starter (JSON echo, health endpoint, or input transformation). Switching templates confirms before replacing edited source. Supply a valid name, source, positive version, permissions and visibility. Client validation checks the supported name pattern and the 256 KiB source limit; the host remains authoritative.

Creating saves an unpublished function and navigates to its detail page. Enabling public visibility requires confirmation; no endpoint is published by creation. Use the detail workspace to preflight and actually execute the current source against JSON before publishing.

Requires the Functions installation to be ready and enabled. See the [admin route index](README.md) and [HTTP API](../http-api.md).
