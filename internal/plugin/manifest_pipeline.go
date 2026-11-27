package plugin

import "strings"

// PipelineSpec is a bounded sequence of allowlisted helper invocations.
type PipelineSpec struct {
	Steps  []PipelineStep `yaml:"steps" json:"steps"`
	Return string         `yaml:"return" json:"return"`
}

type PipelineStep struct {
	Helper string `yaml:"helper" json:"helper"`
	Input  string `yaml:"input" json:"input"`
	Save   string `yaml:"save" json:"save"`
}

func authReference(value string) bool {
	switch value {
	case "$auth.actor.id", "$auth.actor.kind", "$auth.actor.authenticated", "$auth.actor.level",
		"$auth.caller.id", "$auth.caller.kind", "$auth.caller.authenticated", "$auth.caller.level":
		return true
	default:
		return false
	}
}

func pipelineReference(value string, available map[string]bool) bool {
	if value == "$project.id" || value == "$plugin.id" || authReference(value) {
		return true
	}
	parts := strings.Split(value, ".")
	if len(parts) > 8 {
		return false
	}
	switch parts[0] {
	case "$input":
	case "$steps":
		if len(parts) < 2 || !available[parts[1]] {
			return false
		}
	default:
		return false
	}
	for _, key := range parts[1:] {
		if !validResponseKey(key) {
			return false
		}
	}
	return true
}

func (s *PipelineSpec) valid() bool {
	if s == nil || s.Steps == nil || len(s.Steps) > 32 {
		return false
	}
	names := map[string]bool{}
	for _, step := range s.Steps {
		switch step.Helper {
		case "string.trim", "string.nonempty", "string.lower", "string.upper", "value.identity", "number.integer", "boolean.require", "observe.fire":
		default:
			return false
		}
		if !validResponseKey(step.Save) || names[step.Save] || !pipelineReference(step.Input, names) {
			return false
		}
		names[step.Save] = true
	}
	return pipelineReference(s.Return, names)
}
