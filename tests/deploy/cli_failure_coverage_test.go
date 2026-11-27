package deploy_test

import (
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCLIRejectsBrokenPluginDiscoveryAndSandboxPrerequisites(t *testing.T) {
	binary := buildBrainCLI(t)
	env := []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture", "PROJECTS_FILE=brain.sqlite"}
	const manifest = "id: fixture\nversion: '1'\nhost_api: 1\nentrypoint: index.ts\nsource_files:\n  - index.ts\n"
	for _, tc := range []struct {
		name, manifest, source, expected string
		duplicate                        bool
	}{
		{"missing source", manifest, "", "sources rejected", false},
		{"duplicate", manifest, "export {}", "duplicate plugin", true},
		{"incompatible", strings.Replace(manifest, "host_api: 1", "host_api: 2", 1), "export {}", "rejected", false},
		{"missing dependency", manifest + "dependencies:\n  - other@1\n", "export {}", "plugin discovery:", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			names := []string{"first"}
			if tc.duplicate {
				names = append(names, "second")
			}
			for _, name := range names {
				root := filepath.Join(dir, "plugins", name)
				if err := os.MkdirAll(root, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "plugin.yaml"), []byte(tc.manifest), 0600); err != nil {
					t.Fatal(err)
				}
				if tc.source != "" {
					if err := os.WriteFile(filepath.Join(root, "index.ts"), []byte(tc.source), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			runBrainCLI(t, binary, dir, nil, env, tc.expected, false)
		})
	}
	dir := t.TempDir()
	root := filepath.Join(dir, "plugins", "fixture")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "plugin.yaml"), []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.ts"), []byte("export {}"), 0600); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"bwrap", "node"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	env = append(env, "PATH="+bin)
	runBrainCLI(t, binary, dir, nil, env, "loader unavailable", false)
	loader := filepath.Join(dir, "node_modules", "tsx", "dist")
	if err := os.MkdirAll(loader, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loader, "loader.mjs"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	runBrainCLI(t, binary, dir, nil, env, "sandbox startup probe failed", false)
	runBrainCLI(t, binary, dir, nil, append(env, "TMPDIR="+filepath.Join(dir, "missing")), "no such file or directory", false)
}

func TestHealthOnlyShutdownDoesNotStartProduction(t *testing.T) {
	binary := buildBrainCLI(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	cmd := exec.Command(binary, "health-only")
	cmd.Dir = t.TempDir()
	cmd.Env = []string{"HTTP_ADDRESS=" + address}
	if cover := os.Getenv("BRAIN_CLI_COVERDIR"); cover != "" {
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+cover)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	client := &http.Client{Timeout: 100 * time.Millisecond}
	deadline := time.Now().Add(5 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		res, err := client.Get("http://" + address + "/_system/livez")
		if err == nil {
			res.Body.Close()
			if res.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatal("health-only did not start")
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("health-only shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("health-only did not exit")
	}
}
