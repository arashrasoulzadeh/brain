package execution_test

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"brain/internal/execution"
)

func registration(t *testing.T, projectID, version string) execution.Registration {
	t.Helper()
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "main.ts", Data: []byte("export {}")}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	return execution.Registration{
		Key:         execution.ArtifactKey{ProjectID: projectID, ID: "files", Version: version},
		Runtime:     execution.Runtime{Kind: "typescript", Profile: "sandboxed"},
		Entrypoints: []string{"files.metadata", "files.list"}, Bundle: bundle,
	}
}

func TestCatalogPinsVersionAndScope(t *testing.T) {
	catalog, err := execution.NewCatalog(4, 4096)
	if err != nil {
		t.Fatal(err)
	}
	r := registration(t, "project-a", "1.0.0")
	if err := catalog.Register(r); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Register(r); err != nil {
		t.Fatal("identical registration must be idempotent", err)
	}
	a, runtime, err := catalog.Resolve(r.Key, "files.metadata")
	if err != nil || runtime != r.Runtime || a.Digest != execution.ArtifactDigest(r.Bundle) || a.ProjectID != "project-a" {
		t.Fatalf("bad resolution: %v %v %v", a, runtime, err)
	}
	original := append([]byte(nil), r.Bundle...)
	r.Bundle[0] ^= 1
	r.Entrypoints[0] = "unknown"
	b, err := catalog.Fetch(a)
	if err != nil || string(b) != string(original) {
		t.Fatal("catalog retained mutable input", err)
	}
	b[0] ^= 1
	b, err = catalog.Fetch(a)
	if err != nil || string(b) != string(original) {
		t.Fatal("catalog returned mutable storage", err)
	}
	bad := a
	bad.Digest = "sha256:" + strings.Repeat("a", 64)
	if _, err := catalog.Fetch(bad); !errors.Is(err, execution.ErrArtifactMismatch) {
		t.Fatal("wrong digest accepted")
	}
	bad = a
	bad.Entrypoint = "unknown"
	if _, err := catalog.Fetch(bad); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatal("unknown entrypoint accepted")
	}
	bad = a
	bad.ProjectID = "project-b"
	if _, err := catalog.Fetch(bad); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatal("cross-project artifact accepted")
	}
	if _, _, err := catalog.Resolve(r.Key, "unknown"); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatal("unknown entrypoint resolved")
	}
	shared := registration(t, "", "1.0.0")
	if err := catalog.Register(shared); err != nil {
		t.Fatal(err)
	}
	other := r.Key
	other.ProjectID = "project-b"
	if _, _, err := catalog.Resolve(other, "files.metadata"); !errors.Is(err, execution.ErrArtifactNotFound) {
		t.Fatal("unexpected fallback to shared artifact")
	}
	if _, _, err := catalog.Resolve(shared.Key, "files.metadata"); err != nil {
		t.Fatal(err)
	}
	request := validRequest(time.Now())
	request.Artifact.ProjectID = "project-b"
	if !errors.Is(request.Validate(time.Now()), execution.ErrInvalidRequest) {
		t.Fatal("request admitted a foreign artifact scope")
	}
}

func TestCatalogRejectsVersionReplacementAndInvalidRegistration(t *testing.T) {
	catalog, err := execution.NewCatalog(2, 4096)
	if err != nil {
		t.Fatal(err)
	}
	r := registration(t, "", "1.0.0")
	if err := catalog.Register(r); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(*execution.Registration){
		"different bytes": func(v *execution.Registration) {
			v.Bundle, _ = execution.BuildBundle([]execution.BundleFile{{Path: "main.ts", Data: []byte("changed")}}, 4096)
		},
		"different runtime":     func(v *execution.Registration) { v.Runtime.Profile = "trusted" },
		"different entrypoints": func(v *execution.Registration) { v.Entrypoints = []string{"other"} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			v := registration(t, "", "1.0.0")
			mutate(&v)
			if !errors.Is(catalog.Register(v), execution.ErrArtifactConflict) {
				t.Fatal("version replacement accepted")
			}
		})
	}
	invalid := map[string]func(*execution.Registration){
		"missing ID":           func(v *execution.Registration) { v.Key.ID = "" },
		"floating version":     func(v *execution.Registration) { v.Key.Version = "latest" },
		"invalid project":      func(v *execution.Registration) { v.Key.ProjectID = " " },
		"missing runtime":      func(v *execution.Registration) { v.Runtime.Kind = "" },
		"missing profile":      func(v *execution.Registration) { v.Runtime.Profile = "" },
		"missing entrypoints":  func(v *execution.Registration) { v.Entrypoints = nil },
		"blank entrypoint":     func(v *execution.Registration) { v.Entrypoints = []string{" "} },
		"duplicate entrypoint": func(v *execution.Registration) { v.Entrypoints = []string{"one", "one"} },
		"invalid bundle":       func(v *execution.Registration) { v.Bundle = []byte("not a bundle") },
	}
	for name, mutate := range invalid {
		t.Run(name, func(t *testing.T) {
			v := registration(t, "", "1.0.0")
			mutate(&v)
			if err := catalog.Register(v); err == nil {
				t.Fatal("invalid registration accepted")
			}
		})
	}
	if err := catalog.Register(registration(t, "", "2.0.0")); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(catalog.Register(registration(t, "", "3.0.0")), execution.ErrCatalogLimit) {
		t.Fatal("entry limit ignored")
	}
	for _, limits := range [][2]int64{{0, 100}, {1, 0}} {
		if _, err := execution.NewCatalog(int(limits[0]), limits[1]); err == nil {
			t.Fatal("invalid limits accepted")
		}
	}
	tiny, err := execution.NewCatalog(4, int64(len(r.Bundle)))
	if err != nil || tiny.Register(r) != nil {
		t.Fatal("initial registration failed")
	}
	if !errors.Is(tiny.Register(registration(t, "", "2.0.0")), execution.ErrCatalogLimit) {
		t.Fatal("byte limit ignored")
	}
	tooLarge := registration(t, "", "2.0.0")
	tooLarge.Bundle = append(tooLarge.Bundle, 0)
	if !errors.Is(tiny.Register(tooLarge), execution.ErrCatalogLimit) {
		t.Fatal("oversized registration accepted")
	}
}

func TestCatalogConcurrentResolution(t *testing.T) {
	catalog, err := execution.NewCatalog(1, 4096)
	if err != nil {
		t.Fatal(err)
	}
	r := registration(t, "", "1.0.0")
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := catalog.Register(r); err != nil {
				t.Error(err)
				return
			}
			a, _, err := catalog.Resolve(r.Key, "files.list")
			if err != nil {
				t.Error(err)
				return
			}
			if _, err := catalog.Fetch(a); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
