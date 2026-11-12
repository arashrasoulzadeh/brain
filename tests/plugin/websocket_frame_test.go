package plugin_test

import (
	"testing"

	"brain/internal/plugin"
)

func TestWebSocketFrameSizeIsBounded(t *testing.T) {
	if err := plugin.ValidateWebSocketFrame(make([]byte, 1024), 1024); err != nil {
		t.Fatalf("frame at limit rejected: %v", err)
	}
	if err := plugin.ValidateWebSocketFrame(make([]byte, 1025), 1024); err == nil {
		t.Fatal("oversized frame accepted")
	}
}
