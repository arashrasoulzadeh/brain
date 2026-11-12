package plugin_test

import (
	"errors"
	"testing"
	"time"

	"brain/internal/plugin"
)

func TestIsolationPolicyRequiresRestrictedWorkerDefaults(t *testing.T) {
	policy := plugin.IsolationPolicy{FilesystemReadOnly: true, NetworkEnabled: false, CPUMillis: 100, MemoryBytes: 64 << 20, ProcessLimit: 16, WallTime: time.Second}
	if err := policy.Validate(); err != nil {
		t.Fatal(err)
	}
	policy.NetworkEnabled = true
	if !errors.Is(policy.Validate(), plugin.ErrInvalidIsolationPolicy) {
		t.Fatal("network-enabled policy accepted")
	}
}
