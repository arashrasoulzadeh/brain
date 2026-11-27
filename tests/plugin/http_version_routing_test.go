package plugin_test

import (
	"testing"

	"brain/internal/plugin"
)

func TestPromotedProjectHTTPVersionAndRollback(t *testing.T) {
	registry := plugin.NewRegistry(1)
	for _, version := range []string{"1.0.0", "2.0.0"} {
		m := plugin.Manifest{ID: "rooms", Version: version, HostAPI: 1, Mount: "/rooms", Capabilities: []string{"rooms.read"},
			Routes: []plugin.Route{{Method: "GET", Path: "/", Auth: "operator", Operation: "rooms.list", Function: "listRooms", Capability: "rooms.read"}}}
		if err := registry.Register(m); err != nil {
			t.Fatal(err)
		}
		if err := registry.Install("alpha", m.ID, version); err != nil {
			t.Fatal(err)
		}
		if err := registry.Grant("alpha", m.ID, version, "rooms.read"); err != nil {
			t.Fatal(err)
		}
		if err := registry.SetReadiness(m.ID, version, true); err != nil {
			t.Fatal(err)
		}
		if err := registry.Enable("alpha", m.ID, version); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms"); ok {
		t.Fatal("unpromoted overlapping versions exposed")
	}
	if err := registry.Promote("alpha", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if route, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms"); !ok || route.Version != "1.0.0" {
		t.Fatalf("promoted version 1: %#v, %v", route, ok)
	}
	if err := registry.Promote("alpha", "rooms", "2.0.0"); err != nil {
		t.Fatal(err)
	}
	if route, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms"); !ok || route.Version != "2.0.0" {
		t.Fatalf("promoted version 2: %#v, %v", route, ok)
	}
	if err := registry.Rollback("alpha", "rooms", "1.0.0"); err != nil {
		t.Fatal(err)
	}
	if route, ok := registry.MatchHTTPRoute("alpha", "GET", "/rooms"); !ok || route.Version != "1.0.0" {
		t.Fatalf("rollback version 1: %#v, %v", route, ok)
	}
	if _, ok := registry.MatchHTTPRoute("beta", "GET", "/rooms"); ok {
		t.Fatal("version leaked across projects")
	}
}
