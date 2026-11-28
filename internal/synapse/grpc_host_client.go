package synapse

import (
	"context"
	"encoding/json"

	"brain/internal/execution"
	pb "brain/proto/synapse/v1"
)

type hostReply struct {
	output json.RawMessage
	err    error
}

func (c *remoteNeuron) callHost(ctx context.Context, job *remoteJob, operation string, input json.RawMessage) (json.RawMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.jobs[job.lease.Attempt] != job || job.hostReply != nil || job.hostCallID >= execution.MaxHostCalls ||
		!execution.ValidHostCall(operation, input) || job.request.Artifact.Entrypoint != "functions.invoke" ||
		job.request.Runtime.Kind != execution.UserFunctionRuntimeKind {
		c.mu.Unlock()
		return nil, execution.ErrHostCallDenied
	}
	job.hostCallID++
	callID := job.hostCallID
	reply := make(chan hostReply, 1)
	job.hostReply = reply
	job.hostReceived = false
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		job.hostReply = nil
		c.mu.Unlock()
	}()
	message := &pb.HostRequest{Parent: leaseWire(job.lease), CallId: callID, Operation: &pb.HostRequest_Execute{Execute: &pb.NestedInvocation{Operation: operation, InputJson: input}}}
	if err := c.send(&pb.NeuronMessage{Body: &pb.NeuronMessage_HostRequest{HostRequest: message}}); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-reply:
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return result.output, result.err
	}
}

func (c *remoteNeuron) receiveHostResponse(response *pb.HostResponse) error {
	if response == nil {
		return ErrProtocol
	}
	lease, err := readLease(response.Parent)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if lease.SessionID != c.session || lease.Attempt > c.lastAttempt {
		return ErrProtocol
	}
	job := c.jobs[lease.Attempt]
	if job == nil {
		return nil // A response can cross terminal cancellation.
	}
	if job.lease != lease || response.CallId == 0 || response.CallId != job.hostCallID || job.hostReceived {
		return ErrProtocol
	}
	if job.hostReply == nil {
		return nil // A cancelled local call may already have stopped waiting.
	}
	var reply hostReply
	switch value := response.Outcome.(type) {
	case *pb.HostResponse_OutputJson:
		if len(value.OutputJson) > execution.MaxHostCallBytes || !json.Valid(value.OutputJson) {
			return ErrProtocol
		}
		reply.output = value.OutputJson
	case *pb.HostResponse_Error:
		if value.Error == nil || value.Error.Code != "host_call_denied" {
			return ErrProtocol
		}
		reply.err = execution.ErrHostCallDenied
	default:
		return ErrProtocol
	}
	select {
	case job.hostReply <- reply:
		job.hostReceived = true
		return nil
	default:
		return ErrProtocol
	}
}
