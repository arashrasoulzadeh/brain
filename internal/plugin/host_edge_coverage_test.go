package plugin

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCoverageDiscoveryBrokenManifests(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "sample")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(root); !errors.Is(err, ErrPluginManifestMissing) {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "plugin.yaml")
	if err := os.Symlink("plugin.yaml", path); err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(root); err == nil || errors.Is(err, ErrPluginManifestMissing) {
		t.Fatalf("symlink loop was not reported: %v", err)
	}
	for _, name := range []string{"", "/absolute", "a\\b", "a\x00b", "../escape"} {
		if validRelativeManifestPath(name) {
			t.Fatalf("accepted invalid source path %q", name)
		}
	}
}

func TestCoverageMissingRegistryAndObserverGrants(t *testing.T) {
	r := coverageLifecycleRegistry()
	req := WorkerRequest{Protocol: 1, ProjectID: "coverage", PluginID: "sample", Version: "1", Operation: "read"}
	delete(r.registered, "sample@1")
	if err := r.StartWorker("coverage", "sample", "1"); !errors.Is(err, ErrNotRegistered) {
		t.Fatal(err)
	}
	if err := r.authorizeInvocation(req, false); !errors.Is(err, ErrGrantDenied) {
		t.Fatal(err)
	}
	manifest := Manifest{ID: "sample", Version: "1", Capabilities: []string{"read"}, Observers: []ObserverSpec{{Plugin: "rooms", Event: "connected", Function: "observe", Capability: "read"}}}
	r.registered["sample@1"] = Registered{Manifest: manifest}
	req.Operation = manifest.Observers[0].operation()
	install := r.installations["coverage:sample@1"]
	install.Draining = true
	install.DrainUntil = time.Now().Add(time.Hour)
	r.installations["coverage:sample@1"] = install
	if err := r.authorizeInvocation(req, true); !errors.Is(err, ErrDisabled) {
		t.Fatal(err)
	}
	install.Draining = false
	install.Resources = map[string]map[string]bool{"read": {"another.event": true}}
	r.installations["coverage:sample@1"] = install
	if err := r.authorizeInvocation(req, true); !errors.Is(err, ErrGrantDenied) {
		t.Fatal(err)
	}
	req.Payload = []byte(`{"projectId":"other"}`)
	if _, err := r.invokeUsingWorker(req, false, nil, nil); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	if err := r.CheckAdminReadAsOperator(req, "alice"); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
}

func TestCoverageCredentialBoundaryRejections(t *testing.T) {
	r, _, req := coverageDelegationRegistry()
	now := time.Now()
	call := HostCall{ProjectID: req.ProjectID, PluginID: req.PluginID, Version: req.Version, Operation: req.Operation, Capability: "read", ActorID: "alice"}
	broker := &Broker{Registry: r, ActorAllows: func(string, string, string) bool { return true }}
	credential, err := broker.Issue(call, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	bad := req
	bad.Protocol = 2
	if _, err := InvokeWithCredential(broker, credential, nil, bad, 1, now); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	if _, err := r.InvokeWithCredential(broker, credential, bad, now); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatal(err)
	}
	if _, err := InvokeWithCredential(broker, credential, nil, req, 1, now); !errors.Is(err, ErrWorkerNotReady) {
		t.Fatal(err)
	}
	if _, err := InvokeWithCredential(broker, credential, nil, req, 1, now); !errors.Is(err, ErrInvocationExpired) {
		t.Fatal(err)
	}
	credential, err = broker.Issue(call, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	manifest := r.registered["target@1"].Manifest
	manifest.Routes[0].ResourceField = "id"
	r.registered["target@1"] = Registered{Manifest: manifest}
	if _, err := InvokeWithCredential(broker, credential, coverageStopFailureWorker{}, req, 1, now); !errors.Is(err, ErrInvocationCredentialMismatch) {
		t.Fatal(err)
	}
	credential = InvocationCredential{Call: call}
	r.installations["coverage:access-management@1"] = Installation{ProjectID: "coverage", PluginID: accessPluginID, Version: "1"}
	if _, err := InvokeWithCredential(broker, credential, nil, req, 1, now); !errors.Is(err, ErrInvocationCredentialMismatch) {
		t.Fatal(err)
	}
}

func TestCoverageStorageCorruptionAndDerivedPage(t *testing.T) {
	s := coverageStorage(t)
	if err := s.PutState("", "registry", "lifecycle", map[string]any{"installations": "bad"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.LoadRegistryState(); err == nil {
		t.Fatal("corrupt lifecycle snapshot accepted")
	}
	r := NewRegistry(1)
	if err := r.BindSQLiteStorage(s); err == nil {
		t.Fatal("corrupt registry storage bound")
	}
	manifest := Manifest{Models: map[string]ModelSpec{"record": {Version: 1}}, Admin: &AdminSpec{Version: 1, Pages: []AdminPage{{ID: adminModelPageID("record")}}}}
	pages := manifest.AdminForOperator()
	if pages == nil || len(pages.Pages) != 1 || pages.Pages[0].Model != nil {
		t.Fatal("explicit page was overridden")
	}
	if err := r.ApproveArtifact("missing", "1", "http://invalid"); !errors.Is(err, ErrNotRegistered) {
		t.Fatal(err)
	}
}
