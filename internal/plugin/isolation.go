package plugin

import (
	"errors"
	"time"
)

var ErrInvalidIsolationPolicy = errors.New("invalid worker isolation policy")

type IsolationPolicy struct {
	FilesystemReadOnly bool
	NetworkEnabled     bool
	CPUMillis          int
	MemoryBytes        int64
	ProcessLimit       int
	WallTime           time.Duration
}

func (p IsolationPolicy) Validate() error {
	if !p.FilesystemReadOnly || p.NetworkEnabled || p.CPUMillis <= 0 || p.MemoryBytes <= 0 || p.ProcessLimit <= 0 || p.WallTime <= 0 {
		return ErrInvalidIsolationPolicy
	}
	return nil
}
