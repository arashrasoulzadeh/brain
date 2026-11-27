package plugin_test

import (
	"crypto/rand"
	"errors"
	"testing"
	"testing/iotest"
	"time"

	"brain/internal/plugin"
)

func TestBrokerCredentialIssuanceFailsClosed(t *testing.T) {
	now := time.Unix(99, 0)
	call := plugin.HostCall{ProjectID: "alpha", ActorID: "alice", PluginID: "records", Version: "1", Capability: "read", Operation: "read", ExpiresAt: now.Add(time.Minute)}
	var absent *plugin.Broker
	if err := absent.Authorize(call, now); !errors.Is(err, plugin.ErrRegistryUnavailable) {
		t.Fatalf("nil broker: %v", err)
	}
	b := plugin.Broker{ActorAllows: func(string, string, string) bool { return true }}
	if err := b.Authorize(call, now); !errors.Is(err, plugin.ErrRegistryUnavailable) {
		t.Fatalf("missing registry: %v", err)
	}
	if err := b.CommitCredential(plugin.InvocationCredential{}, now); !errors.Is(err, plugin.ErrInvocationExpired) {
		t.Fatalf("empty credential: %v", err)
	}
	if _, err := b.Issue(call, now, 0); !errors.Is(err, plugin.ErrInvocationExpired) {
		t.Fatalf("zero lifetime: %v", err)
	}
	if _, err := b.Issue(call, now, time.Minute); !errors.Is(err, plugin.ErrRegistryUnavailable) {
		t.Fatalf("unauthorized issuance: %v", err)
	}
	b.Registry = plugin.NewRegistry(1)
	if err := b.Authorize(call, now); !errors.Is(err, plugin.ErrNotRegistered) {
		t.Fatalf("unregistered operation: %v", err)
	}
	m := plugin.Manifest{ID: "records", Version: "1", HostAPI: 1, Capabilities: []string{"read"}, Routes: []plugin.Route{{Operation: "read", Function: "read", Capability: "read"}}}
	if err := b.Registry.Register(m); err != nil {
		t.Fatal(err)
	}
	installAccessTestPlugin(t, b.Registry, "alpha", m, &contractWorker{ready: true})
	failure := errors.New("entropy unavailable")
	original := rand.Reader
	defer func() { rand.Reader = original }()
	rand.Reader = iotest.ErrReader(failure)
	if credential, err := b.Issue(call, now, time.Minute); !errors.Is(err, failure) || credential.Token != "" {
		t.Fatalf("credential issued without entropy: %+v %v", credential, err)
	}
}
