package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestWebSocketMessageRequiresProjectRoomAndProtocol(t *testing.T) {
	message := plugin.WebSocketMessage{Protocol: 1, RequestID: "req-1", ProjectID: "dev", RoomID: "room-1", Type: "room.join"}
	if err := message.Validate(1); err != nil {
		t.Fatal(err)
	}
	message.RoomID = ""
	if !errors.Is(message.Validate(1), plugin.ErrInvalidWebSocketMessage) {
		t.Fatal("roomless message accepted")
	}
}

func TestWebSocketMessageRejectsHostProjectMismatch(t *testing.T) {
	message := plugin.WebSocketMessage{Protocol: 1, RequestID: "req-1", ProjectID: "dev", RoomID: "room-1", Type: "input"}
	if err := message.ValidateForProject(1, "prod"); !errors.Is(err, plugin.ErrInvalidWebSocketMessage) {
		t.Fatalf("cross-project message accepted: %v", err)
	}
	if err := message.ValidateForProject(1, "dev"); err != nil {
		t.Fatal(err)
	}
}
