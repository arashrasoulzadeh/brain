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
