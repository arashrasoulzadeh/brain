package plugin

// ExecutionBoundary describes the implementation placement of a declared
// route. It is not a permission grant and does not select a runtime version.
type ExecutionBoundary string

const (
	BoundaryInvalid    ExecutionBoundary = ""
	BoundaryNeuron     ExecutionBoundary = "neuron"
	BoundaryStorage    ExecutionBoundary = "host-storage"
	BoundaryDelegation ExecutionBoundary = "host-delegation"
	BoundaryRecovery   ExecutionBoundary = "host-recovery"
)

// ExecutionBoundary centralizes the helper allowlist shared by manifest
// validation, scheduled admission and the Neuron's request decoder.
// Host helpers never become executable sandbox entrypoints.
func (r Route) ExecutionBoundary() ExecutionBoundary {
	switch r.Helper {
	case "", "sdk.dispatch", "sdk.request", "sdk.response", "sdk.context", "sdk.pipeline":
		return BoundaryNeuron
	case "host.model":
		return BoundaryStorage
	case "host.call":
		return BoundaryDelegation
	case "host.admin":
		return BoundaryRecovery
	default:
		return BoundaryInvalid
	}
}
