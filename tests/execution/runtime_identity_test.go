package execution_test

import (
	"os"
	"path/filepath"
	"testing"

	"brain/internal/execution"
)

func TestRuntimeSourceDigestPinsRuntimeFiles(t *testing.T) {
	root := t.TempDir()
	names := []string{"package.json", "package-lock.json", "tsconfig.json", "scripts/typescript-worker.mjs", "scripts/function-worker.mjs", "scripts/run-sandboxed-typescript-worker.sh", "plugins/sdk-typescript/src/index.ts"}
	for _, name := range names {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	first, err := execution.RuntimeSourceDigest(root)
	if err != nil {
		t.Fatal(err)
	}
	again, err := execution.RuntimeSourceDigest(root)
	if err != nil || again != first {
		t.Fatalf("unstable digest: %s %v", again, err)
	}
	sdk := filepath.Join(root, "plugins/sdk-typescript/src/index.ts")
	if err := os.WriteFile(sdk, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	changed, err := execution.RuntimeSourceDigest(root)
	if err != nil || changed == first {
		t.Fatalf("changed SDK did not change digest: %s %v", changed, err)
	}
	if err := os.Remove(sdk); err != nil {
		t.Fatal(err)
	}
	if _, err := execution.RuntimeSourceDigest(root); err == nil {
		t.Fatal("missing SDK entrypoint accepted")
	}
	if err := os.Symlink(filepath.Join(root, "package.json"), sdk); err != nil {
		t.Fatal(err)
	}
	if _, err := execution.RuntimeSourceDigest(root); err == nil {
		t.Fatal("symlink accepted")
	}
	if err := os.Remove(filepath.Join(root, "package-lock.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := execution.RuntimeSourceDigest(root); err == nil {
		t.Fatal("missing package lock accepted")
	}
}
