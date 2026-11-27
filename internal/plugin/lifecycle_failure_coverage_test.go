package plugin

import (
	"errors"
	"testing"
)

type coverageStopFailureWorker struct{ failure error }

func (w coverageStopFailureWorker) Start() error { return nil }
func (w coverageStopFailureWorker) Stop() error  { return w.failure }
func (w coverageStopFailureWorker) Ready() bool  { return true }
func (w coverageStopFailureWorker) Handle(req WorkerRequest) WorkerResponse {
	return WorkerResponse{Protocol: req.Protocol, Success: true}
}

func coverageLifecycleRegistry() *Registry {
	r := NewRegistry(1)
	r.registered["sample@1"] = Registered{Manifest: Manifest{ID: "sample", Version: "1", Capabilities: []string{"read"}}}
	r.installations["coverage:sample@1"] = Installation{ProjectID: "coverage", PluginID: "sample", Version: "1", Enabled: true, Grants: map[string]bool{"read": true}}
	r.readiness["sample@1"] = true
	return r
}

func TestCoverageLifecycleStopFailure(t *testing.T) {
	failure := errors.New("injected worker stop failure")
	for _, operation := range []string{"disable", "revoke", "uninstall"} {
		t.Run(operation, func(t *testing.T) {
			r := coverageLifecycleRegistry()
			if err := r.BindWorker("coverage", "sample", "1", coverageStopFailureWorker{failure}); err != nil {
				t.Fatal(err)
			}
			var err error
			switch operation {
			case "disable":
				var preview DisablePreview
				preview, err = r.PreviewDisable("coverage", "sample", "1")
				if err != nil {
					t.Fatal(err)
				}
				err = r.ApplyDisablePreview(preview)
			case "revoke":
				err = r.Revoke("coverage", "sample", "1", "read")
			case "uninstall":
				err = r.UninstallProject("coverage")
			}
			if !errors.Is(err, failure) {
				t.Fatalf("stop failure not reported: %v", err)
			}
			if r.workers["coverage:sample@1"] != nil {
				t.Fatal("failed worker still bound")
			}
			if err := r.Authorize("coverage", "sample", "1", "read"); err == nil {
				t.Fatal("authorization retained after failed stop")
			}
		})
	}
}

func TestCoverageLifecyclePersistenceFailure(t *testing.T) {
	for _, operation := range []string{"remove", "uninstall"} {
		t.Run(operation, func(t *testing.T) {
			r := coverageLifecycleRegistry()
			s := coverageStorage(t)
			if err := r.BindSQLiteStorage(s); err != nil {
				t.Fatal(err)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			var err error
			if operation == "remove" {
				err = r.Remove("coverage", "sample", "1")
			} else {
				err = r.UninstallProject("coverage")
			}
			if !errors.Is(err, ErrLifecycleStorage) {
				t.Fatalf("storage failure not reported: %v", err)
			}
			if err := r.Authorize("coverage", "sample", "1", "read"); !errors.Is(err, ErrLifecycleStorage) {
				t.Fatal(err)
			}
		})
	}
}

func TestCoverageLifecycleReferencesAndGrants(t *testing.T) {
	r := coverageLifecycleRegistry()
	r.active["coverage:sample"] = "1"
	if err := r.UninstallProject("coverage"); !errors.Is(err, ErrActiveWork) {
		t.Fatal(err)
	}
	delete(r.installations, "coverage:sample@1")
	if err := r.Unregister("sample", "1"); !errors.Is(err, ErrVersionReferenced) {
		t.Fatal(err)
	}
	delete(r.active, "coverage:sample")
	r.pins["empty"] = ""
	r.pins["work"] = "coverage:sample@1"
	if err := r.Unregister("sample", "1"); !errors.Is(err, ErrVersionReferenced) {
		t.Fatal(err)
	}
	if version, ok := r.PinnedVersion("work"); ok || version != "" {
		t.Fatal("missing installation resolved")
	}
	r = coverageLifecycleRegistry()
	r.installations["coverage:sample@1"] = Installation{ProjectID: "coverage", PluginID: "sample", Version: "1"}
	if err := r.Grant("coverage", "sample", "1", "read"); err != nil {
		t.Fatal(err)
	}
	if err := r.Promote("coverage", "sample", "1"); !errors.Is(err, ErrDisabled) {
		t.Fatal(err)
	}
	if err := r.Revoke("other", "sample", "1", "read"); !errors.Is(err, ErrNotInstalled) {
		t.Fatal(err)
	}
	if err := r.Authorize("!", "sample", "1", "read"); !errors.Is(err, ErrProjectInvalid) {
		t.Fatal(err)
	}
	if configTypeMatches("unknown", "value") {
		t.Fatal("unknown configuration type accepted")
	}
	r.readiness["sample@1"] = false
	if err := r.Enable("coverage", "sample", "1"); !errors.Is(err, ErrNotReady) {
		t.Fatal(err)
	}
}
