package synapse

import (
	"context"
	"encoding/json"
	"strings"

	"brain/internal/execution"
)

// BeginHostCall admits at most one outstanding call and sixteen total calls
// per live, accepted operator draft lease. It never trusts a worker's project,
// principal, or requested credentials. The caller must execute asynchronously
// so the stream receiver remains available to process child results.
func (s *Session) BeginHostCall(lease Lease, callID uint64, operation string, input json.RawMessage) (context.Context, execution.HostCallFunc, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.lookupLocked(lease)
	if p == nil || !p.accepted || p.cancelling || p.ctx.Err() != nil || p.hostBusy ||
		callID != p.hostCalls+1 || callID > execution.MaxHostCalls || !execution.ValidHostCall(operation, input) {
		return nil, nil, ErrProtocol
	}
	req := p.assignment.Request
	if req.Runtime.Kind != execution.UserFunctionRuntimeKind || req.Artifact.Entrypoint != "functions.invoke" ||
		req.Auth.Capability != "functions.invoke" || req.ParentID != "" || req.Auth.OriginalCaller != req.Auth.Caller ||
		!strings.HasPrefix(req.Auth.Caller, "operator:") || len(req.Auth.Caller) <= len("operator:") {
		return nil, nil, ErrProtocol
	}
	handler := execution.HostHandler(p.ctx)
	if handler == nil {
		return nil, nil, ErrProtocol
	}
	p.hostCalls, p.hostBusy = callID, true
	return execution.WithParent(p.ctx, req), handler, nil
}

// FinishHostCall fences late responses against the original lease. A child
// whose parent has completed, cancelled, or disconnected cannot return data.
func (s *Session) FinishHostCall(lease Lease, callID uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.lookupLocked(lease)
	if p == nil || !p.hostBusy || p.hostCalls != callID {
		return false
	}
	p.hostBusy = false
	return !p.cancelling && p.ctx.Err() == nil
}
