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
