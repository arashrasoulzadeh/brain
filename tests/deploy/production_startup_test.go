package deploy_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func buildBrainCLI(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "brain")
	args := []string{"build", "-buildvcs=false", "-o", binary}
	if os.Getenv("BRAIN_CLI_COVERDIR") != "" {
		args = append(args, "-cover", "-covermode=atomic", "-coverpkg=./internal/...,./cmd/...")
	}
	build := exec.Command("go", append(args, "./cmd/brain")...)
	build.Dir = "../.."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build host: %v\n%s", err, output)
	}
	return binary
}

func runBrainCLI(t *testing.T, binary, dir string, args, env []string, want string, success bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = dir
	command.Env = append([]string{"HTTP_ADDRESS=127.0.0.1:0"}, env...)
	if cover := os.Getenv("BRAIN_CLI_COVERDIR"); cover != "" {
		command.Env = append(command.Env, "GOCOVERDIR="+cover)
	}
	output, err := command.CombinedOutput()
	if (err == nil) != success || ctx.Err() != nil || !strings.Contains(string(output), want) {
		t.Fatalf("CLI %v: success=%t, want %t; error=%v, output=%s", args, err == nil, success, err, output)
	}
}

func TestProductionBinaryRejectsUnsafeStartupConfiguration(t *testing.T) {
	binary := buildBrainCLI(t)
	for _, tc := range []struct {
		name string
		env  []string
		want string
	}{
		{name: "missing operator", want: "OPERATOR_USERNAME and OPERATOR_SECRET are required"},
		{name: "missing storage", env: []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture"}, want: "PROJECTS_FILE must name an explicit durable SQLite file"},
		{name: "memory storage", env: []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture", "PROJECTS_FILE=:memory:"}, want: "PROJECTS_FILE must name an explicit durable SQLite file"},
		{name: "unsupported database", env: []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture", "RUNTIME_DSN=postgres://fixture", "PROJECTS_FILE=brain.sqlite"}, want: "refusing implicit SQLite fallback"},
		{name: "blank address", env: []string{"HTTP_ADDRESS="}, want: "HTTP_ADDRESS is required"},
		{name: "invalid operator", env: []string{"OPERATOR_USERNAME= ", "OPERATOR_SECRET=fixture", "PROJECTS_FILE=brain.sqlite"}, want: "operator bootstrap failed"},
		{name: "missing discovery root", env: []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture", "PROJECTS_FILE=brain.sqlite"}, want: "plugin discovery unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runBrainCLI(t, binary, t.TempDir(), nil, tc.env, tc.want, false)
		})
	}
}

func TestProductionCLICommandsAndDiscoveryErrors(t *testing.T) {
	binary := buildBrainCLI(t)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(root, "plugins", "dashboard", "plugin.yaml")
	for _, tc := range []struct {
		args    []string
		env     []string
		want    string
		success bool
	}{
		{[]string{"plugin-contract"}, nil, "usage: brain plugin-contract", false},
		{[]string{"plugin-contract", "missing.yaml", "markdown"}, nil, "missing.yaml", false},
		{[]string{"plugin-contract", manifest, "markdown"}, nil, "dashboard", true},
		{[]string{"plugin-contract", manifest, "typescript"}, nil, "export interface Operations", true},
		{[]string{"plugin-contract", manifest, "unknown"}, nil, "unknown contract format", false},
		{[]string{"bootstrap-operator"}, nil, "bootstrap requires", false},
		{[]string{"bootstrap-operator"}, []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture"}, "bootstrapped", true},
		{[]string{"bootstrap-dashboard"}, nil, "dashboard plugin dashboard@", true},
		{[]string{"verify-object-manifest", "missing.json"}, nil, "missing.json", false},
	} {
		runBrainCLI(t, binary, t.TempDir(), tc.args, tc.env, tc.want, tc.success)
	}
	dir := t.TempDir()
	objects := filepath.Join(dir, "objects.json")
	if err := os.WriteFile(objects, []byte(`{"objects":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	runBrainCLI(t, binary, dir, []string{"verify-object-manifest", objects}, nil, "object manifest verified", true)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	runBrainCLI(t, binary, dir, []string{"health-only"}, []string{"HTTP_ADDRESS=" + listener.Addr().String()}, "address already in use", false)

	env := []string{"OPERATOR_USERNAME=ci", "OPERATOR_SECRET=fixture", "PROJECTS_FILE=brain.sqlite"}
	plugins := filepath.Join(dir, "plugins")
	if err := os.Mkdir(plugins, 0o700); err != nil {
		t.Fatal(err)
	}
	runBrainCLI(t, binary, dir, nil, env, "no plugins discovered", false)
	for i, source := range []string{"invalid: [", "id: invalid\n"} {
		location := filepath.Join(plugins, fmt.Sprintf("fixture-%d", i))
		if err := os.Mkdir(location, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(location, "plugin.yaml"), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
		runBrainCLI(t, binary, dir, nil, env, "rejected", false)
		if err := os.RemoveAll(location); err != nil {
			t.Fatal(err)
		}
	}
	// An empty PATH reliably makes the mandatory sandbox runtime unavailable.
	runBrainCLI(t, binary, root, nil, append(env, "PATH="+t.TempDir()), "sandbox runtime requires", false)
}
