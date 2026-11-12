package plugin_test

import (
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestHeartbeatTrackerAcceptsFreshPingAndExpiresIdleConnection(t *testing.T) {
	now := time.Unix(100, 0)
	tracker := plugin.NewHeartbeatTracker(10 * time.Second)
	tracker.Touch(now)
	if !tracker.Alive(now.Add(9 * time.Second)) {
		t.Fatal("fresh heartbeat marked dead")
	}
	if tracker.Alive(now.Add(10 * time.Second)) {
		t.Fatal("idle heartbeat remained alive")
	}
}
