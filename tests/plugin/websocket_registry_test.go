package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

type trackedConnection struct {
	closed bool
}

func (c *trackedConnection) Close() error {
	c.closed = true
	return nil
}

func TestWebSocketRegistryClosesAndRemovesConnections(t *testing.T) {
	registry := plugin.NewWebSocketRegistry()
	first := &trackedConnection{}
	second := &trackedConnection{}
	registry.Register("first", first)
	registry.Register("second", second)

	if err := registry.CloseAll(); err != nil {
		t.Fatal(err)
	}
	if !first.closed || !second.closed {
		t.Fatal("not all WebSocket connections were closed")
	}
	if registry.Count() != 0 {
		t.Fatalf("connection count = %d, want 0", registry.Count())
	}
	if errors.Is(registry.CloseAll(), plugin.ErrWebSocketConnectionStillTracked) {
		t.Fatal("closed connections remained tracked")
	}
}

func TestWebSocketRegistryClosesOnlyProjectConnections(t *testing.T) {
	registry := plugin.NewWebSocketRegistry()
	dev := &trackedConnection{}
	prod := &trackedConnection{}
	if err := registry.RegisterProject("dev-1", "dev", dev); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterProject("prod-1", "prod", prod); err != nil {
		t.Fatal(err)
	}
	if err := registry.CloseProject("dev"); err != nil {
		t.Fatal(err)
	}
	if !dev.closed || prod.closed {
		t.Fatalf("project close state = dev:%v prod:%v", dev.closed, prod.closed)
	}
	if registry.Count() != 1 {
		t.Fatalf("connection count = %d, want 1", registry.Count())
	}
}

func TestWebSocketRegistryRejectsProjectlessRegistration(t *testing.T) {
	if err := plugin.NewWebSocketRegistry().RegisterProject("connection", "", &trackedConnection{}); !errors.Is(err, plugin.ErrWebSocketProjectRequired) {
		t.Fatalf("projectless registration error = %v", err)
	}
}
