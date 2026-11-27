package main

import (
	"testing"
)

func TestSandboxRuntimePlatformPolicy(t *testing.T) {
	for _, tc := range []struct {
		platform, local, address string
		launcher, binary         string
	}{
		{"linux", "", "0.0.0.0:8080", "scripts/run-sandboxed-typescript-worker.sh", "bwrap"},
		{"darwin", "1", "127.0.0.1:8080", "scripts/run-macos-typescript-worker.sh", "/usr/bin/sandbox-exec"},
		{"darwin", "", "127.0.0.1:8080", "", ""},
		{"darwin", "1", "0.0.0.0:8080", "", ""},
	} {
		launcher, binary, factory, err := sandboxRuntime(tc.platform, tc.local, tc.address)
		if tc.launcher == "" {
			if err == nil || factory != nil {
				t.Fatalf("unsafe platform policy accepted: %+v", tc)
			}
			continue
		}
		if err != nil || factory == nil || launcher != tc.launcher || binary != tc.binary {
			t.Fatalf("wrong platform configuration: %q %q %v", launcher, binary, err)
		}
	}
}
