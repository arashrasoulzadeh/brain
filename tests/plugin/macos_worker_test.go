package plugin_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestMacOSWorkerIsolation(t *testing.T) {
	if runtime.GOOS != "darwin" || os.Getenv("BRAIN_TEST_MACOS_SANDBOX") != "1" {
		t.Skip("requires macOS and installed Node dependencies")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	secret := filepath.Join(directory, "secret.txt")
	if err := os.WriteFile(secret, []byte("private fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	source := `import { readFileSync, writeFileSync } from "node:fs";
import { connect } from "node:net";
export async function health(input: { secret: string }) {
  let readDenied = false, writeDenied = false;
  try { readFileSync(input.secret); } catch { readDenied = true; }
  try { writeFileSync(input.secret, "changed"); } catch { writeDenied = true; }
  const networkDenied = await new Promise<boolean>(resolve => {
    const socket = connect({host:"127.0.0.1",port:18093});
    socket.on("connect", () => {socket.destroy(); resolve(false);});
    socket.on("error", err => {socket.destroy(); resolve(["EPERM","EACCES"].includes((err as NodeJS.ErrnoException).code ?? ""));});
    socket.setTimeout(500, () => {socket.destroy(); resolve(false);});
  });
  return { readDenied, writeDenied, networkDenied };
}`
	if err := os.WriteFile(filepath.Join(directory, "index.mts"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := plugin.Manifest{ID: "macos-probe", Version: "1.0.0", HostAPI: 1, Entry: "index.mts", SourceFiles: []string{"index.mts"}, Imports: []string{"node:fs", "node:net"}, Capabilities: []string{"health"}, Routes: []plugin.Route{{Operation: "health", Function: "health", Capability: "health"}}}
	policy := plugin.IsolationPolicy{FilesystemReadOnly: true, CPUMillis: 30000, MemoryBytes: 2 << 30, ProcessLimit: 512, WallTime: 2 * time.Second}
	worker, err := plugin.NewManifestMacOSTypeScriptWorker(filepath.Join(root, "scripts/run-macos-typescript-worker.sh"), filepath.Join(root, "scripts/typescript-worker.mjs"), filepath.Join(directory, "descriptor.json"), directory, "dev", manifest, policy)
	if err != nil {
		t.Fatal(err)
	}
	worker.SetTrustedStartupDiagnostics(os.Stderr)
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer worker.Stop()
	payload, _ := json.Marshal(map[string]string{"secret": secret})
	response, err := plugin.InvokeForProject(worker, plugin.WorkerRequest{Protocol: 1, ProjectID: "dev", PluginID: manifest.ID, Version: manifest.Version, Operation: "health", Payload: payload}, 1, "dev")
	if err != nil || !response.Success || string(response.Payload) != `{"readDenied":true,"writeDenied":true,"networkDenied":true}` {
		t.Fatalf("macOS sandbox probe: %+v %v", response, err)
	}
}
