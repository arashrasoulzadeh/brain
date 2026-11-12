package plugin_test

import (
	"testing"

	"brain/internal/plugin"
)

func TestWebSocketOriginMustBeAllowlisted(t *testing.T) {
	if !plugin.AllowedWebSocketOrigin("https://app.example", []string{"https://app.example"}) {
		t.Fatal("allowlisted origin rejected")
	}
	if plugin.AllowedWebSocketOrigin("https://evil.example", []string{"https://app.example"}) {
		t.Fatal("unallowlisted origin accepted")
	}
}
