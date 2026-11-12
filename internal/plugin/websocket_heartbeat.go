package plugin

import "time"

type HeartbeatTracker struct {
	timeout time.Duration
	last    time.Time
}

func NewHeartbeatTracker(timeout time.Duration) *HeartbeatTracker {
	if timeout <= 0 {
		panic("heartbeat timeout must be positive")
	}
	return &HeartbeatTracker{timeout: timeout}
}

func (h *HeartbeatTracker) Touch(now time.Time) { h.last = now }

func (h *HeartbeatTracker) Alive(now time.Time) bool {
	return !h.last.IsZero() && now.Sub(h.last) < h.timeout
}
