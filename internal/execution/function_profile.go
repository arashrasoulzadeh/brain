package execution

import "encoding/json"

// FunctionProfile is fixed-size diagnostic metadata, not user output or proof
// of resource enforcement. All durations use local monotonic clocks.
type FunctionProfile struct {
	ArtifactValidationNanos int64 `json:"artifact_validation_ns"`
	SandboxPreparationNanos int64 `json:"sandbox_preparation_ns"`
	SandboxStartupNanos int64 `json:"sandbox_startup_ns"`
	SandboxInvocationNanos int64 `json:"sandbox_invocation_ns"`
	SandboxStopNanos int64 `json:"sandbox_stop_ns"`
	FilesystemCleanupNanos int64 `json:"filesystem_cleanup_ns"`
	ArtifactFetchNanos int64 `json:"artifact_fetch_ns,omitempty"`
	BundleOpenNanos int64 `json:"bundle_open_ns,omitempty"`
	NeuronDurationNanos int64 `json:"neuron_duration_ns,omitempty"`
}

// AddFunctionProfile preserves the existing function envelope and its raw
// result. If metadata cannot be added, retain the original valid output.
func AddFunctionProfile(output json.RawMessage, profile FunctionProfile) json.RawMessage {
	var envelope map[string]json.RawMessage
	if json.Unmarshal(output, &envelope) != nil || envelope == nil || envelope["result"] == nil {
		return output
	}
	raw, err := json.Marshal(profile)
	if err != nil {
		return output
	}
	envelope["profile"] = raw
	encoded, err := json.Marshal(envelope)
	if err != nil || len(encoded) > 1<<20 {
		return output
	}
	return encoded
}

// ReadFunctionProfile accepts only the fixed numeric schema, bounded to the
// invocation time scale. It never accepts arbitrary labels or text.
func ReadFunctionProfile(output json.RawMessage) (FunctionProfile, bool) {
	var envelope struct {
		Profile *FunctionProfile `json:"profile"`
	}
	if json.Unmarshal(output, &envelope) != nil || envelope.Profile == nil {
		return FunctionProfile{}, false
	}
	p := *envelope.Profile
	for _, v := range []int64{p.ArtifactValidationNanos,p.SandboxPreparationNanos,p.SandboxStartupNanos,p.SandboxInvocationNanos,p.SandboxStopNanos,p.FilesystemCleanupNanos,p.ArtifactFetchNanos,p.BundleOpenNanos,p.NeuronDurationNanos} {
		if v < 0 || v > 60_000_000_000 {
			return FunctionProfile{}, false
		}
	}
	return p, true
}

type FunctionStage struct {
	Name string
	Nanos int64
}

// Stages is a fixed allowlist shared by execution logs and metrics.
func (p FunctionProfile) Stages() []FunctionStage {
	return []FunctionStage{
		{"artifact fetch", p.ArtifactFetchNanos},
		{"bundle verification", p.BundleOpenNanos},
		{"runtime artifact validation", p.ArtifactValidationNanos},
		{"sandbox preparation", p.SandboxPreparationNanos},
		{"sandbox startup", p.SandboxStartupNanos},
		{"sandbox invocation", p.SandboxInvocationNanos},
		{"sandbox termination", p.SandboxStopNanos},
		{"filesystem cleanup", p.FilesystemCleanupNanos},
	}
}
