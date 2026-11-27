package plugin_test

import (
	"brain/internal/plugin"
	"testing"
)

func TestManifestHTTPRouteAdmission(t *testing.T) {
	valid := plugin.Manifest{ID: "rooms", Entry: "index.ts", SourceFiles: []string{"index.ts"}, Mount: "/rooms",
		Capabilities: []string{"rooms.read"}, Routes: []plugin.Route{{Operation: "rooms.list", Function: "listRooms", Capability: "rooms.read", Method: "GET", Path: "/", Auth: "operator"}}}
	if err := plugin.ValidateManifestContract(valid); err != nil {
		t.Fatalf("valid HTTP route: %v", err)
	}
	for name, mutate := range map[string]func(*plugin.Manifest){
		"system mount":          func(m *plugin.Manifest) { m.Mount = "/_system" },
		"project UI mount":      func(m *plugin.Manifest) { m.Mount = "/ui" },
		"path traversal":        func(m *plugin.Manifest) { m.Routes[0].Path = "/../secret" },
		"encoded slash":         func(m *plugin.Manifest) { m.Routes[0].Path = "/%2fsecret" },
		"unsupported method":    func(m *plugin.Manifest) { m.Routes[0].Method = "TRACE" },
		"missing auth":          func(m *plugin.Manifest) { m.Routes[0].Auth = "" },
		"invalid binding":       func(m *plugin.Manifest) { m.Routes[0].Function = "../../worker" },
		"undeclared capability": func(m *plugin.Manifest) { m.Routes[0].Capability = "other.read" },
		"duplicate template": func(m *plugin.Manifest) {
			m.Routes = append(m.Routes, plugin.Route{Operation: "rooms.get", Function: "getRoom", Capability: "rooms.read", Method: "GET", Path: "/", Auth: "operator"})
		},
	} {
		t.Run(name, func(t *testing.T) {
			m := valid
			m.Routes = append([]plugin.Route(nil), valid.Routes...)
			mutate(&m)
			if err := plugin.ValidateManifestContract(m); err == nil {
				t.Fatal("invalid HTTP route accepted")
			}
		})
	}
}
