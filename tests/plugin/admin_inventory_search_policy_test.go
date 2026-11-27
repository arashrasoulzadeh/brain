package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

type inventoryWithoutSearch struct{}

func (inventoryWithoutSearch) Supports(string) bool { return true }
func (inventoryWithoutSearch) Prepare(request plugin.WorkerRequest) (plugin.WorkerRequest, func() error, error) {
	return request, nil, nil
}
func (inventoryWithoutSearch) AdminInventory(_, _, _, _, _ string, _ int) (plugin.AdminInventoryPage, error) {
	panic("unfiltered enumeration must not substitute for search")
}
func (inventoryWithoutSearch) AdminInventoryScoped(_, _, _, _, _ string, _ int, _ plugin.AdminInventoryFilter) (plugin.AdminInventoryPage, error) {
	panic("unfiltered enumeration must not substitute for search")
}

func TestAdminInventorySearchEnforcesExactOperatorScope(t *testing.T) {
	m, err := plugin.LoadManifest("../../plugins/files/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	r, storage := accessRegistry(t, m)
	for _, project := range []string{"alpha", "beta"} {
		installAccessTestPlugin(t, r, project, accessManifest(t), &contractWorker{ready: true})
		installAccessTestPlugin(t, r, project, m, &contractWorker{ready: true})
	}
	for _, key := range []string{"match-a", "match-b", "match-c", "other"} {
		if err := storage.PutFile("alpha", key, "content", false); err != nil {
			t.Fatal(err)
		}
	}
	if err := storage.PutFile("beta", "match-foreign", "private", false); err != nil {
		t.Fatal(err)
	}
	route, _ := m.Route("files.metadata")
	allowRevision := createAccessRule(t, r, "alpha", "search-allow", "operator:root", m, route, "*", true)
	createAccessRule(t, r, "alpha", "search-deny", "operator:root", m, route, "match-b", false)
	for _, key := range []string{"match-a", "match-b", "match-c"} {
		if err := r.GrantResource("alpha", m.ID, m.Version, route.Capability, key); err != nil {
			t.Fatal(err)
		}
	}
	search := func(actor, project, after string) plugin.AdminInventoryPage {
		t.Helper()
		page, err := r.AdminInventorySearchAsOperator(actor, project, m.ID, m.Version, "operations", after, " MATCH ", 1)
		if err != nil {
			t.Fatal(err)
		}
		return page
	}
	first := search("root", "alpha", "")
	if len(first.Rows) != 1 || first.Rows[0].Key != "match-a" || first.Next != "match-a" {
		t.Fatalf("filtered first page: %+v", first)
	}
	second := search("root", "alpha", first.Next)
	if len(second.Rows) != 1 || second.Rows[0].Key != "match-c" || second.Next != "" {
		t.Fatalf("filtered second page: %+v", second)
	}
	for _, project := range []string{"alpha", "beta"} {
		if page := search("other", project, ""); len(page.Rows) != 0 || page.Next != "" {
			t.Fatalf("search crossed actor or project scope: %+v", page)
		}
	}
	if _, err := r.AdminInventorySearchAsOperator(" bad", "alpha", m.ID, m.Version, "operations", "", "match", 1); !errors.Is(err, plugin.ErrActorDenied) {
		t.Fatalf("invalid actor: %v", err)
	}
	revokeAccessRule(t, r, "alpha", "search-allow", allowRevision)
	if page := search("root", "alpha", ""); len(page.Rows) != 0 || page.Next != "" {
		t.Fatalf("revoked search exposed rows: %+v", page)
	}
	r.BindInvocationStorage(inventoryWithoutSearch{})
	if _, err := r.AdminInventorySearchAsOperator("root", "alpha", m.ID, m.Version, "operations", "", "match", 1); !errors.Is(err, plugin.ErrAdminInventoryUnavailable) {
		t.Fatalf("managed provider without search support: %v", err)
	}
	installAccessTestPlugin(t, r, "unmanaged", m, &contractWorker{ready: true})
	if _, err := r.AdminInventorySearchAsOperator("root", "unmanaged", m.ID, m.Version, "operations", "", "match", 1); !errors.Is(err, plugin.ErrAdminInventoryUnavailable) {
		t.Fatalf("unmanaged provider without search support: %v", err)
	}
}
