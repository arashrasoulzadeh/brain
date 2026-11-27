package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "operations.d.ts")
	if err := generate(filepath.Join(root, "missing"), output); err == nil {
		t.Fatal("missing manifests accepted")
	}
	if err := generate(root, filepath.Join(root, "missing", "operations.d.ts")); err == nil {
		t.Fatal("missing output directory accepted")
	}
	if err := generate(root, output); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil || !strings.Contains(string(data), "export interface ManifestOperations") {
		t.Fatalf("generated declarations: %s (%v)", data, err)
	}
}

func TestManifestSDKCommand(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "manifest-sdk")
	args := []string{"build", "-buildvcs=false", "-o", binary}
	cover := os.Getenv("BRAIN_CLI_COVERDIR")
	if cover != "" {
		args = append(args, "-cover", "-covermode=atomic")
	}
	args = append(args, ".")
	build := exec.Command("go", args...)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build generator: %s (%v)", output, err)
	}
	root := t.TempDir()
	run := func() ([]byte, error) {
		command := exec.Command(binary)
		command.Dir = root
		command.Env = os.Environ()
		if cover != "" {
			command.Env = append(command.Env, "GOCOVERDIR="+cover)
		}
		return command.CombinedOutput()
	}
	if _, err := run(); err == nil {
		t.Fatal("command accepted missing plugin directory")
	}
	dir := filepath.Join(root, "plugins", "sdk-typescript", "src")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if output, err := run(); err != nil {
		t.Fatalf("generate declarations: %s (%v)", output, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "manifest-operations.generated.d.ts")); err != nil {
		t.Fatal(err)
	}
}
