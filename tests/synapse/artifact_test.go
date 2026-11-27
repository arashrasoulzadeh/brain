package synapse_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"brain/internal/execution"
	"brain/internal/synapse"
)

func TestArtifactTransfer(t *testing.T) {
	s := newSession(t, 1, 2, 0, 0)
	ctx := context.Background()
	catalog, err := execution.NewCatalog(1, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	data, err := execution.BuildBundle([]execution.BundleFile{{Path: "main.ts", Data: bytes.Repeat([]byte("x"), synapse.ArtifactChunkBytes)}}, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	r := req("artifact")
	key := execution.ArtifactKey{ID: r.Artifact.ID, Version: r.Artifact.Version}
	if err = catalog.Register(execution.Registration{Key: key, Runtime: r.Runtime, Entrypoints: []string{r.Artifact.Entrypoint}, Bundle: data}); err != nil {
		t.Fatal(err)
	}
	r.Artifact, _, err = catalog.Resolve(key, r.Artifact.Entrypoint)
	if err != nil {
		t.Fatal(err)
	}
	done := submit(s, ctx, r)
	assignment := next(t, s).Assignment
	lease := assignment.Lease
	a, err := synapse.NewAssembler(lease, r.Artifact, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	chunks := 0
	if err = synapse.StreamArtifact(ctx, s, catalog, lease, func(chunk synapse.Chunk) error { chunks++; return a.Push(chunk) }); err != nil || chunks < 2 {
		t.Fatal(chunks, err)
	}
	out, err := a.Finish()
	if err != nil || !bytes.Equal(out, data) {
		t.Fatal(err)
	}
	if _, err = a.Finish(); !errors.Is(err, synapse.ErrInvalidChunk) {
		t.Fatal(err)
	}
	failed := errors.New("send failure")
	if err = synapse.StreamArtifact(ctx, s, catalog, lease, func(synapse.Chunk) error { return failed }); !errors.Is(err, failed) {
		t.Fatal(err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if err = synapse.StreamArtifact(cancelCtx, s, catalog, lease, func(synapse.Chunk) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err = synapse.StreamArtifact(ctx, nil, catalog, lease, func(synapse.Chunk) error { return nil }); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	empty, err := execution.NewCatalog(1, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if err = synapse.StreamArtifact(ctx, s, empty, lease, func(synapse.Chunk) error { return nil }); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatal(err)
	}
	if err = s.Accept(lease); err != nil {
		t.Fatal(err)
	}
	if err = synapse.StreamArtifact(ctx, s, catalog, lease, func(synapse.Chunk) error { return s.Complete(lease, success(r.ID)) }); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
	if v := receive(t, done); v.err != nil {
		t.Fatal(v)
	}
	if err = synapse.StreamArtifact(ctx, s, catalog, lease, func(synapse.Chunk) error { return nil }); !errors.Is(err, synapse.ErrProtocol) {
		t.Fatal(err)
	}
}
func TestChunkValidation(t *testing.T) {
	lease := synapse.Lease{SessionID: "session", Attempt: 1, ExecutionID: "id"}
	data := []byte("abc")
	artifact := execution.Artifact{Digest: execution.ArtifactDigest(data)}
	if _, err := synapse.NewAssembler(synapse.Lease{}, artifact, 128); !errors.Is(err, synapse.ErrInvalidChunk) {
		t.Fatal(err)
	}
	if _, err := synapse.NewAssembler(lease, artifact, 0); !errors.Is(err, synapse.ErrInvalidChunk) {
		t.Fatal(err)
	}
	valid := synapse.Chunk{Lease: lease, Total: 3, Data: data, Final: true}
	for name, mutate := range map[string]func(*synapse.Chunk){"wrong lease": func(c *synapse.Chunk) { c.Lease.Attempt++ }, "offset": func(c *synapse.Chunk) { c.Offset = 1 }, "invalid total": func(c *synapse.Chunk) { c.Total = 0 }, "oversized": func(c *synapse.Chunk) { c.Total = 129 }, "empty": func(c *synapse.Chunk) { c.Data = nil }, "too much data": func(c *synapse.Chunk) { c.Total = 2 }, "missing final": func(c *synapse.Chunk) { c.Final = false }, "premature final": func(c *synapse.Chunk) { c.Total = 4 }} {
		t.Run(name, func(t *testing.T) {
			a, err := synapse.NewAssembler(lease, artifact, 128)
			if err != nil {
				t.Fatal(err)
			}
			c := valid
			mutate(&c)
			if err = a.Push(c); !errors.Is(err, synapse.ErrInvalidChunk) {
				t.Fatal(err)
			}
			if err = a.Push(valid); !errors.Is(err, synapse.ErrInvalidChunk) {
				t.Fatal("poisoned transfer recovered", err)
			}
		})
	}
	a, err := synapse.NewAssembler(lease, artifact, 128)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Finish(); !errors.Is(err, synapse.ErrInvalidChunk) {
		t.Fatal(err)
	}
	bad := valid
	bad.Data = []byte("abd")
	if err = a.Push(bad); !errors.Is(err, execution.ErrArtifactMismatch) {
		t.Fatal(err)
	}
	a, err = synapse.NewAssembler(lease, artifact, 128)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Push(valid); err != nil {
		t.Fatal(err)
	}
	if err = a.Push(valid); !errors.Is(err, synapse.ErrInvalidChunk) {
		t.Fatal(err)
	}
	a, err = synapse.NewAssembler(lease, artifact, 128)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Push(synapse.Chunk{Lease: lease, Total: 3, Data: []byte("a")}); err != nil {
		t.Fatal(err)
	}
	if err = a.Push(synapse.Chunk{Lease: lease, Offset: 1, Total: 4, Data: []byte("bcd"), Final: true}); !errors.Is(err, synapse.ErrInvalidChunk) {
		t.Fatal(err)
	}
}
