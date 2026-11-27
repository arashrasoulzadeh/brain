package plugin_test

import (
	"brain/internal/plugin"
	"testing"
)

func TestProjectPluginHTTPRoutesFollowLifecycle(t *testing.T) {
	registry := plugin.NewRegistry(1)
	m := plugin.Manifest{ID: "games", Version: "1", HostAPI: 1, Mount: "/games", Capabilities: []string{"games.read"},
		Routes: []plugin.Route{{Operation: "games.modes", Function: "modes", Capability: "games.read", Method: "GET", Path: "/", Auth: "operator"}}}
	if err := registry.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := registry.Install("alpha", "games", "1"); err != nil {
		t.Fatal(err)
	}
	if err := registry.Grant("alpha", "games", "1", "games.read"); err != nil {
		t.Fatal(err)
	}
	if err := registry.SetReadiness("games", "1", true); err != nil {
		t.Fatal(err)
	}
	if _, ok := registry.MatchHTTPRoute("alpha", "GET", "/games"); ok {
		t.Fatal("disabled route exposed")
	}
	if err := registry.Enable("alpha", "games", "1"); err != nil {
		t.Fatal(err)
	}
	if match, ok := registry.MatchHTTPRoute("alpha", "GET", "/games"); !ok || match.Operation != "games.modes" {
		t.Fatalf("enabled route = %#v, %v", match, ok)
	}
	if _, ok := registry.MatchHTTPRoute("beta", "GET", "/games"); ok {
		t.Fatal("cross-project route exposed")
	}
	if _, ok := registry.MatchHTTPRoute("alpha", "POST", "/games"); ok {
		t.Fatal("undeclared method exposed")
	}
	preview, err := registry.PreviewDisable("alpha", "games", "1")
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.ApplyDisablePreview(preview); err != nil {
		t.Fatal(err)
	}
	if _, ok := registry.MatchHTTPRoute("alpha", "GET", "/games"); ok {
		t.Fatal("disabled route remains exposed")
	}
}
