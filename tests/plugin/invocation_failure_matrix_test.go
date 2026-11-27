package plugin_test

import (
	"crypto/rand"
	"errors"
	"testing"
	"testing/iotest"

	"brain/internal/plugin"
)

func TestInvocationDeniesChangedResourceAndContextBeforeCommit(t *testing.T) {
	failure := errors.New("project unavailable")
	for _, name := range []string{"resource", "entropy", "revoked during preparation", "resolve failed", "project inactive at admission", "project inactive after worker", "protocol", "other grant", "commit failure", "lifecycle unhealthy"} {
		t.Run(name, func(t *testing.T) {
			r := plugin.NewRegistry(1)
			m := plugin.Manifest{ID: "sample", Version: "1", HostAPI: 1, Capabilities: []string{"read", "write"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read", ResourceField: "key"}}}
			if err := r.Register(m); err != nil {
				t.Fatal(err)
			}
			installAccessTestPlugin(t, r, "alpha", m, &contractWorker{ready: true})
			req := plugin.WorkerRequest{Protocol: 1, ProjectID: "alpha", PluginID: m.ID, Version: m.Version, Operation: "read", Payload: []byte(`{"key":"one"}`)}
			calls, commits, resolves := 0, 0, 0
			r.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
				resolves++
				if name == "resolve failed" || (name == "project inactive after worker" && calls > 0) {
					return plugin.ProjectSnapshot{}, failure
				}
				if name == "project inactive at admission" && resolves > 1 {
					return plugin.ProjectSnapshot{ID: id, Active: false}, nil
				}
				return plugin.ProjectSnapshot{ID: id, Active: true}, nil
			})
			worker := &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
				calls++
				protocol := req.Protocol
				if name == "protocol" {
					protocol++
				}
				return plugin.WorkerResponse{Protocol: protocol, Success: true}
			}}
			r.BindInvocationStorage(testStoragePlanner{prepare: func(req plugin.WorkerRequest) (plugin.WorkerRequest, func() error, error) {
				if name == "resource" {
					req.Payload = []byte(`{"key":"other"}`)
				}
				if name == "revoked during preparation" {
					if err := r.Revoke("alpha", m.ID, m.Version, "read"); err != nil {
						t.Fatal(err)
					}
				}
				return req, func() error { commits++; return failure }, nil
			}})
			var want error
			wantCalls, wantCommits := 0, 0
			switch name {
			case "resource":
				want = plugin.ErrInvalidWorkerRequest
			case "entropy":
				original := rand.Reader
				defer func() { rand.Reader = original }()
				rand.Reader = iotest.ErrReader(failure)
				want = failure
			case "revoked during preparation", "other grant":
				want = plugin.ErrGrantDenied
			case "resolve failed":
				want = failure
			case "project inactive at admission":
				want = plugin.ErrProjectInvalid
			case "project inactive after worker":
				want = failure
				wantCalls = 1
			case "protocol":
				want = plugin.ErrInvalidWorkerRequest
				wantCalls = 1
			case "commit failure":
				want = failure
				wantCalls, wantCommits = 1, 1
			case "lifecycle unhealthy":
				store := &coverageStateStore{fail: true}
				if err := r.BindRegistryStorage(store); err != nil {
					t.Fatal(err)
				}
				if err := r.FlushLifecycleState(); !errors.Is(err, plugin.ErrLifecycleStorage) {
					t.Fatal(err)
				}
				want = plugin.ErrLifecycleStorage
			}
			if name == "other grant" {
				if err := r.Revoke("alpha", m.ID, m.Version, "write"); err != nil {
					t.Fatal(err)
				}
			}
			if err := r.BindWorker("alpha", m.ID, m.Version, worker); err != nil {
				t.Fatal(err)
			}
			if _, err := r.Invoke(req); !errors.Is(err, want) {
				t.Fatalf("invocation: %v, want %v", err, want)
			}
			if calls != wantCalls || commits != wantCommits {
				t.Fatalf("worker=%d commit=%d, want %d/%d", calls, commits, wantCalls, wantCommits)
			}
		})
	}
}
