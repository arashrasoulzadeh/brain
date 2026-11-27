package admin_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"brain/internal/admin"
	"brain/internal/operator"
)

func TestFunctionRejectsScalarAndMalformedRuntimeOutput(t *testing.T) {
	service := admin.NewLifecycleService(nil, nil)
	actor := operator.Actor{Kind: operator.ActorOperator, ID: "fixture"}
	const prefix = "export const config = { method: 'POST' }; export default () => "
	for _, expression := range []string{"1", "true", "'scalar'"} {
		if _, err := service.RunFunctionDraft(actor, "result-coverage", "scalar", prefix+expression, nil); err == nil || !strings.Contains(err.Error(), "JSON object, array, or null") {
			t.Fatalf("scalar result accepted: %v", err)
		}
	}
	for _, expression := range []string{"null", "[1,2]", "({valid:true})"} {
		if _, err := service.RunFunctionDraft(actor, "result-coverage", "valid", prefix+expression, nil); err != nil {
			t.Fatalf("valid result rejected: %v", err)
		}
	}
	for _, output := range []string{"not JSON", `{}`, `{"result":1e999}`} {
		t.Run(output, func(t *testing.T) {
			dir := t.TempDir()
			// Simulate malformed output from a trusted runtime executable, never from a plugin.
			if err := os.WriteFile(filepath.Join(dir, "node"), []byte("#!/bin/sh\nprintf '%s' '"+output+"'\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir)
			if _, err := service.RunFunctionDraft(actor, "result-coverage", "malformed", prefix+"null", json.RawMessage(`null`)); err == nil || !strings.Contains(err.Error(), "JSON object, array, or null") {
				t.Fatalf("malformed runtime output accepted: %v", err)
			}
		})
	}
	t.Setenv("PATH", t.TempDir())
	if _, err := service.RunFunctionDraft(actor, "result-coverage", "unavailable", prefix+"null", nil); err == nil || !strings.Contains(err.Error(), "function execution failed") {
		t.Fatalf("missing runtime accepted: %v", err)
	}
}
