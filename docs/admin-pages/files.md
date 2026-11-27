# Files

Route: `/{project}/ui/files`

Runs project-scoped object storage operations through the Files plugin. Requires Files to be ready and enabled. Uses the shared admin sidebar.

Select an authorized file key from inventory, inspect size and publication state, or use the visible Upload file, Download file and Publish file actions. Choose a local file or enter text manually. The local-file picker accepts up to 40 KiB and fills the object key from the filename if the key is blank. Upload requires confirmation; replacing an existing key resets its publication state. The inventory refreshes after a successful mutation.

Selected files are encoded as `data:application/octet-stream;base64,...` strings through the existing Files storage contract. Download file offers Save original file for a recognized, valid base64 representation, or Save stored text for the literal stored string. Downloads use a browser attachment, not an inline executable preview. The inventory and metadata report stored representation size, not decoded file size. Direct API callers still receive the stored string. The 64 KiB admin JSON request limit applies, including key and envelope overhead; large objects need a separate bounded binary storage/streaming API.

Do not treat the published flag as a public CDN URL. Every action still uses the authenticated project, enabled version, declared operation and resource grants.

See the [admin route index](README.md) and [HTTP API](../http-api.md).
