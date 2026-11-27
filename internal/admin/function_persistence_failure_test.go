package admin

import (
	"crypto/rand"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

type coverageFunctionEntropyFailure struct{}

func (coverageFunctionEntropyFailure) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestCoverageFunctionRequestIDFallback(t *testing.T) {
	original := rand.Reader
	rand.Reader = coverageFunctionEntropyFailure{}
	t.Cleanup(func() { rand.Reader = original })
	id := createFunctionRequestID()
	if !strings.HasPrefix(id, "function-") || id == "function-" {
		t.Fatalf("invalid fallback identifier: %q", id)
	}
}

func TestCoverageFunctionDraftPersistenceRollback(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Error(err)
		}
	})
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	previous := functionDrafts.items
	functionDrafts.items = make(map[string]FunctionDraft)
	defer func() { functionDrafts.items = previous }()
	if err := os.WriteFile(".local", []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	draft := FunctionDraft{ProjectID: "coverage", Name: "sample", Version: 1, CreatedAt: time.Now().UTC()}
	if err := saveFunctionDraftLocked(draft); err == nil {
		t.Fatal("draft persisted into invalid directory")
	}
	if len(functionDrafts.items) != 0 {
		t.Fatal("failed draft creation retained")
	}
	if err := os.Remove(".local"); err != nil {
		t.Fatal(err)
	}
	if err := saveFunctionDraftLocked(draft); err != nil {
		t.Fatal(err)
	}
	invalid := draft
	invalid.CreatedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := saveFunctionDraftLocked(invalid); err == nil {
		t.Fatal("invalid draft timestamp persisted")
	}
	if got := functionDrafts.items["coverage:sample"]; !got.CreatedAt.Equal(draft.CreatedAt) {
		t.Fatal("failed update replaced last admitted draft")
	}
}
