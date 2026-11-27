# Files

## Developer

Upload with `files.upload` and payload `{"key":"example.txt","data":"hello"}`. Read via `files.metadata` or `files.download`; publish explicitly with `files.publish`. The resource identity is the key. Never treat a client-supplied published flag, stored data or claimed ownership as authorization; production storage planning is host controlled.

## Administrator

Review upload/download/publish grants, allowed object keys, quotas and storage durability. Keep uploads unpublished until the correct authorized publish action. Verify object versions and restore manifests using the [restore runbook](../restore-runbook.md). The in-process worker store is a fixture/fallback, not proof of durable object storage.
