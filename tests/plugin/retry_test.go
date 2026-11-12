package plugin_test

import (
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestRetryPolicyBoundsAttemptsAndSeparatesConflict(t *testing.T) {
	policy := plugin.RetryPolicy{MaxAttempts: 3, BaseDelay: 10 * time.Millisecond}
	if !policy.Retryable(plugin.ErrorUnavailable, 1) || policy.Retryable(plugin.ErrorConflict, 1) || policy.Retryable(plugin.ErrorRetryable, 3) {
		t.Fatal("retry classification incorrect")
	}
	if policy.Delay(3) != 40*time.Millisecond {
		t.Fatalf("delay = %v", policy.Delay(3))
	}
}
