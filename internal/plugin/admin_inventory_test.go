package plugin

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestAdminInventoryScopedPaginatedAndRevocable(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "inventory.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	for _, key := range []string{"a", "b", "c"} {
		if err := storage.PutFile("one", key, "secret content", false); err != nil {
			t.Fatal(err)
		}
	}
	if err := storage.PutFile("two", "other", "private", true); err != nil {
		t.Fatal(err)
	}
	r := NewRegistry(1)
	m := Manifest{ID: "files", Version: "1", HostAPI: 1, Capabilities: []string{"files.metadata"}, Routes: []Route{{Operation: "files.metadata", Function: "metadata", Capability: "files.metadata"}}, Admin: &AdminSpec{Version: 1, Pages: []AdminPage{{ID: "files", Title: "Files", Inventory: &AdminInventorySpec{Kind: "file_keys", Capability: "files.metadata", Title: "Files"}, Actions: []AdminAction{{Title: "Metadata", Operation: "files.metadata"}}}}}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	r.BindInvocationStorage(legacyStoragePlanner{storage: storage})
	if err := r.Install("one", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	if err := r.Grant("one", m.ID, m.Version, "files.metadata"); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	if err := r.Enable("one", m.ID, m.Version); err != nil {
		t.Fatal(err)
	}
	first, err := r.AdminInventory("one", m.ID, m.Version, "files", "", 2)
	if err != nil || len(first.Rows) != 2 || first.Rows[0].Key != "a" || first.Next != "b" || first.Rows[0].SizeBytes == nil || *first.Rows[0].SizeBytes != 14 {
		t.Fatalf("first page: %#v %v", first, err)
	}
	second, err := r.AdminInventory("one", m.ID, m.Version, "files", first.Next, 2)
	if err != nil || len(second.Rows) != 1 || second.Rows[0].Key != "c" || second.Next != "" {
		t.Fatalf("second page: %#v %v", second, err)
	}
	if _, err := r.AdminInventory("two", m.ID, m.Version, "files", "", 2); err == nil {
		t.Fatal("foreign project inventory accepted")
	}
	if _, err := r.AdminInventory("one", m.ID, m.Version, "unknown", "", 2); !errors.Is(err, ErrGrantDenied) {
		t.Fatalf("unknown page: %v", err)
	}
	if err := r.GrantResource("one", m.ID, m.Version, "files.metadata", "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AdminInventory("one", m.ID, m.Version, "files", "", 2); !errors.Is(err, ErrGrantDenied) {
		t.Fatalf("restricted grant allowed enumeration: %v", err)
	}
	if err := r.Revoke("one", m.ID, m.Version, "files.metadata"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AdminInventory("one", m.ID, m.Version, "files", "", 2); !errors.Is(err, ErrGrantDenied) {
		t.Fatalf("revoked grant allowed enumeration: %v", err)
	}
}

func TestLegacyInventoryRejectsNamespaceConfusion(t *testing.T) {
	storage, err := OpenSQLiteStorage(filepath.Join(t.TempDir(), "inventory.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	p := legacyStoragePlanner{storage: storage}
	if _, err := p.AdminInventory("one", "rooms", "file_keys", "rooms.result", "", 10); !errors.Is(err, ErrGrantDenied) {
		t.Fatalf("cross-plugin inventory: %v", err)
	}
	if _, err := p.AdminInventory("one", "files", "file_keys", "files.upload", "", 10); !errors.Is(err, ErrGrantDenied) {
		t.Fatalf("incorrect capability accepted: %v", err)
	}
}
