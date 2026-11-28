package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"brain/internal/app"
)

var errNeuronControl = errors.New("managed Neuron action unavailable")

func launchDevNeuron(id string, cmd *exec.Cmd) (*devNeuronProcess, error) {
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	process := &devNeuronProcess{id: id, cmd: cmd, done: make(chan struct{})}
	go func() {
		process.err = cmd.Wait()
		close(process.done)
	}()
	return process, nil
}

func (p *devNeuronProcess) running() bool {
	select {
	case <-p.done:
		return false
	default:
		return true
	}
}

func (c *devCluster) ManagedNeurons() []app.ManagedNeuron {
	c.processMu.RLock()
	defer c.processMu.RUnlock()
	result := make([]app.ManagedNeuron, 0, len(c.processes))
	for _, p := range c.processes {
		result = append(result, app.ManagedNeuron{ID: p.id, Running: p.running()})
	}
	return result
}

// ControlNeuron can only operate on a process spawned by this local supervisor.
// It drains admission before stopping and never kills a still-active assignment
// to satisfy an operator timeout. A timed-out drain remains paused for recovery.
func (c *devCluster) ControlNeuron(ctx context.Context, id, action string) error {
	if c.controlCtx == nil || !c.controlMu.TryLock() {
		return errNeuronControl
	}
	defer c.controlMu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	stop := context.AfterFunc(c.controlCtx, cancel)
	defer stop()
	if c.controlCtx.Err() != nil || ctx.Err() != nil {
		return errNeuronControl
	}
	c.processMu.RLock()
	index := -1
	var process *devNeuronProcess
	for i, candidate := range c.processes {
		if candidate.id == id {
			index, process = i, candidate
			break
		}
	}
	c.processMu.RUnlock()
	if process == nil {
		return errNeuronControl
	}
	switch action {
	case "drain":
		return c.scheduler.SetOperatorDrainContext(ctx, id, true)
	case "resume":
		if !process.running() {
			return errNeuronControl
		}
		return c.scheduler.SetOperatorDrainContext(ctx, id, false)
	case "start", "restart", "stop":
	default:
		return errNeuronControl
	}
	if action == "start" && process.running() {
		return errNeuronControl
	}
	if err := c.scheduler.SetOperatorDrainContext(ctx, id, true); err != nil {
		return err
	}
	if err := c.scheduler.WaitNeuronIdle(ctx, id); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	stopDevNeuron(process)
	// Wait for the old authenticated connection to disappear; a fresh process
	// cannot replace a live session. The pause survives all reconnects.
	if err := c.waitNeuronConnection(ctx, id, false, nil); err != nil {
		return err
	}
	if action == "stop" {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// The command, arguments, environment and certificate paths are exactly
	// those selected by Brain at startup, never supplied by the request.
	cmd := exec.Command(process.cmd.Path, process.cmd.Args[1:]...)
	cmd.Env = append([]string(nil), process.cmd.Env...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = process.cmd.Dir, process.cmd.Stdout, process.cmd.Stderr
	replacement, err := launchDevNeuron(id, cmd)
	if err != nil {
		return err
	}
	c.processMu.Lock()
	c.processes[index] = replacement
	c.processMu.Unlock()
	if err := c.waitNeuronConnection(ctx, id, true, replacement); err != nil {
		stopDevNeuron(replacement)
		return err
	}
	if c.controlCtx.Err() != nil || ctx.Err() != nil {
		stopDevNeuron(replacement)
		return errNeuronControl
	}
	return c.scheduler.SetOperatorDrainContext(ctx, id, false)
}

func stopDevNeuron(p *devNeuronProcess) {
	if !p.running() {
		return
	}
	_ = p.cmd.Process.Signal(os.Interrupt)
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case <-p.done:
		return
	case <-timer.C:
		_ = p.cmd.Process.Kill()
		<-p.done
	}
}

func (c *devCluster) waitNeuronConnection(ctx context.Context, id string, want bool, process *devNeuronProcess) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if process != nil && !process.running() {
			return errNeuronControl
		}
		if _, connected := c.scheduler.Neuron(id); connected == want {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
