package plugin

import "brain/internal/execution"

// ProcessUsage reports launcher accounting after the process has been reaped.
// It must never read ProcessState concurrently with exec.Cmd.Wait.
func (w *TypeScriptWorker) ProcessUsage() execution.Usage {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.done != nil {
		select {
		case <-w.done:
		default:
			return execution.Usage{}
		}
	}
	state := w.command.ProcessState
	if state == nil {
		return execution.Usage{}
	}
	return execution.Usage{CPUTime: state.UserTime() + state.SystemTime(), PeakMemoryBytes: processPeakRSS(state)}
}
