package plugin_test

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSandboxLauncherFailsClosedWithoutBubblewrap(t *testing.T) {
	launcher, err := filepath.Abs("../../scripts/run-sandboxed-typescript-worker.sh")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(launcher)
	command.Env = []string{"PATH=/usr/bin:/bin", "BWRAP_BIN=missing-bwrap-for-test", "WORKER_SCRIPT=/tmp/worker", "EXECUTION_DESCRIPTOR=/tmp/descriptor", "PLUGIN_ROOT=/tmp", "CPU_MILLIS=1000", "MEMORY_BYTES=65536", "PROCESS_LIMIT=8"}
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("sandbox launcher did not fail closed; output=%s", output)
	}
	if exitError, ok := err.(*exec.ExitError); !ok || exitError.ExitCode() != 78 {
		t.Fatalf("unexpected launcher error: %v", err)
	}
}
