# Files

Scoped content and metadata. [Authoritative manifest](../../plugins/files/plugin.yaml).

## Developer

Upload nonempty string data under a key; read metadata/download through the host-owned persisted snapshot; publish explicitly. Resource grants use the actual key. Do not supply or trust caller-provided published flags or stored content.

## Administrator

The [Files workspace](../admin-pages/files.md) at `/{project}/ui/files` lists authorized keys, metadata and publication state and provides a file picker, manual text input and attachment download. Browser-selected files up to 40 KiB are stored as lossless base64 strings; the API still exposes the stored representation. Grant metadata for inventory and separate upload/download/publish as needed. Review the consequences of publication before confirming.

## Example payloads

files.upload: {"key":"hello.txt","data":"hello"}; files.metadata / files.download / files.publish: {"key":"hello.txt"}.

Payload snippets are arguments to the declared operation, not standalone HTTP routes. Use the [HTTP invocation contract](../http-api.md) and exact manifest version/capability. Host-owned context, internal snapshots and grants must not be supplied by public callers.

## Limitations

Current storage is a SQLite compatibility adapter; not direct cloud object storage, CDN delivery or public signed URL issuance. Publication is a stored flag.
