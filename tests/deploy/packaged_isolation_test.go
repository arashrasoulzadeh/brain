package deploy_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPackagedWorkerRejectsCrossProjectRequest(t *testing.T) {
	probe := buildPackagedIsolationProbe(t)

	for _, test := range []struct {
		name           string
		hostProject    string
		requestProject string
		wantSuccess    bool
	}{
		{name: "project one", hostProject: "project-one", requestProject: "project-one", wantSuccess: true},
		{name: "project two", hostProject: "project-two", requestProject: "project-two", wantSuccess: true},
		{name: "cross project", hostProject: "project-one", requestProject: "project-two", wantSuccess: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := exec.Command(probe, test.hostProject, test.requestProject)
			err := cmd.Run()
			if test.wantSuccess && err != nil {
				t.Fatalf("packaged request failed: %v", err)
			}
			if !test.wantSuccess && err == nil {
				t.Fatal("packaged cross-project request was accepted")
			}
		})
	}
}

func buildPackagedIsolationProbe(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to locate packaged isolation test")
	}
	root := filepath.Join(filepath.Dir(currentFile), "testdata", "packaged-isolation-probe")
	output := filepath.Join(t.TempDir(), "packaged-isolation-probe")
	build := exec.Command("go", "build", "-buildvcs=false", "-trimpath", "-o", output, ".")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build packaged isolation probe: %v\n%s", err, output)
	}
	return output
}
