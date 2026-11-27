package plugin_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"brain/internal/plugin"
)

func ticketCall(project string) plugin.HostCall {
	return plugin.HostCall{
		ProjectID: project, ActorID: project + ":player", PluginID: "rooms", Version: "1",
		Operation: "rooms.connect", Capability: "rooms.connect", Resource: "room-one",
	}
}

func TestWebSocketTicketsBindScopeAndConsumeExactlyOnce(t *testing.T) {
	now := time.Unix(100, 0)
	call := ticketCall("alpha")
	store := plugin.NewWebSocketTicketStore(2, func(candidate plugin.HostCall, at time.Time) error {
		if candidate != (func() plugin.HostCall {
			expected := call
			expected.ExpiresAt = now.Add(time.Second)
			return expected
		})() || at.Before(now) {
			return plugin.ErrActorDenied
		}
		return nil
	})
	token, err := store.Issue(call, now, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Consume("beta", token, now); !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
		t.Fatalf("cross-project redemption: %v", err)
	}
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			binding, err := store.Consume("alpha", token, now)
			if err == nil {
				successes.Add(1)
				if binding.ActorID != call.ActorID || binding.Resource != call.Resource || binding.Version != call.Version || binding.Operation != call.Operation {
					t.Errorf("incorrect binding: %#v", binding)
				}
			} else if !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful redemptions: %d", successes.Load())
	}
}

func TestWebSocketTicketsFailClosedOnExpiryRevocationAndCapacity(t *testing.T) {
	now := time.Unix(200, 0)
	call := ticketCall("alpha")
	var allowed atomic.Bool
	allowed.Store(true)
	store := plugin.NewWebSocketTicketStore(1, func(plugin.HostCall, time.Time) error {
		if !allowed.Load() {
			return plugin.ErrActorDenied
		}
		return nil
	})
	issue := func(at time.Time) string {
		t.Helper()
		token, err := store.Issue(call, at, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	denied := func(token string, at time.Time) {
		t.Helper()
		if _, err := store.Consume("alpha", token, at); !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
			t.Fatalf("redemption: %v", err)
		}
	}
	token := issue(now)
	if _, err := store.Issue(call, now, time.Second); !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
		t.Fatalf("capacity: %v", err)
	}
	denied(token, now.Add(time.Second))
	token = issue(now.Add(time.Second))
	allowed.Store(false)
	denied(token, now.Add(time.Second))
	if _, err := store.Issue(call, now, time.Second); !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
		t.Fatalf("revoked issuance: %v", err)
	}
	allowed.Store(true)
	denied(token, now.Add(time.Second))
	token = issue(now.Add(time.Second))
	store.RevokeProject("beta")
	store.RevokeProject("alpha")
	denied(token, now.Add(time.Second))
	_ = issue(now.Add(time.Second))
	_ = issue(now.Add(2 * time.Second)) // expired entries release capacity
}

func TestWebSocketTicketsRejectInvalidBindings(t *testing.T) {
	now := time.Unix(300, 0)
	allow := func(plugin.HostCall, time.Time) error { return nil }
	store := plugin.NewWebSocketTicketStore(1, allow)
	for _, change := range []func(*plugin.HostCall){
		func(c *plugin.HostCall) { c.ProjectID = "" },
		func(c *plugin.HostCall) { c.ActorID = "beta:player" },
		func(c *plugin.HostCall) { c.ActorID = "alpha:" },
		func(c *plugin.HostCall) { c.PluginID = "" },
		func(c *plugin.HostCall) { c.Version = "" },
		func(c *plugin.HostCall) { c.Operation = "" },
		func(c *plugin.HostCall) { c.Capability = "" },
		func(c *plugin.HostCall) { c.Resource = "" },
		func(c *plugin.HostCall) { c.Resource = "*" },
	} {
		call := ticketCall("alpha")
		change(&call)
		if _, err := store.Issue(call, now, time.Second); !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
			t.Fatalf("invalid binding accepted: %#v", call)
		}
	}
	for _, ttl := range []time.Duration{0, -time.Second, time.Minute + time.Second} {
		if _, err := store.Issue(ticketCall("alpha"), now, ttl); !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
			t.Fatalf("invalid TTL accepted: %v", ttl)
		}
	}
	for _, candidate := range []*plugin.WebSocketTicketStore{nil, plugin.NewWebSocketTicketStore(0, allow), plugin.NewWebSocketTicketStore(1, nil)} {
		if _, err := candidate.Issue(ticketCall("alpha"), now, time.Second); !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
			t.Fatal("invalid store accepted")
		}
	}
	if _, err := store.Consume("alpha", "invalid", now); !errors.Is(err, plugin.ErrWebSocketTicketDenied) {
		t.Fatal("malformed token accepted")
	}
}
