package main

import (
	"os"
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
