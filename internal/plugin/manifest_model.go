package plugin

import "regexp"

var modelNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,47}$`)

type ModelSpec struct {
	Version     int                        `yaml:"version" json:"version"`
	Fields      map[string]string          `yaml:"fields" json:"fields"`
	Initial     string                     `yaml:"initial,omitempty" json:"initial,omitempty"`
	Transitions map[string]ModelTransition `yaml:"transitions,omitempty" json:"transitions,omitempty"`
}

type ModelTransition struct {
	From string `yaml:"from" json:"from"`
	To   string `yaml:"to" json:"to"`
}

type ModelRoute struct {
	Name       string `yaml:"name" json:"name"`
	Action     string `yaml:"action" json:"action"`
	Transition string `yaml:"transition,omitempty" json:"transition,omitempty"`
}

func modelName(value string) bool { return modelNamePattern.MatchString(value) && validResponseKey(value) }

func validModels(models map[string]ModelSpec) bool {
	if len(models) > 32 {
		return false
	}
	for name, spec := range models {
		if !modelName(name) || spec.Version < 1 || len(spec.Fields) == 0 || len(spec.Fields) > 16 || len(spec.Transitions) > 32 {
			return false
		}
		for field, kind := range spec.Fields {
			if !modelName(field) || (kind != "string" && kind != "integer" && kind != "boolean") {
				return false
			}
		}
		if spec.Initial != "" && !modelName(spec.Initial) {
			return false
		}
		if len(spec.Transitions) > 0 && spec.Initial == "" {
			return false
		}
		for action, transition := range spec.Transitions {
			if !modelName(action) || !modelName(transition.From) || !modelName(transition.To) {
				return false
			}
		}
	}
	return true
}

func (route *ModelRoute) valid(models map[string]ModelSpec) bool {
	if route == nil {
		return false
	}
	model, ok := models[route.Name]
	if !ok {
		return false
	}
	switch route.Action {
	case "create", "read", "list", "update", "delete":
		return route.Transition == ""
	case "transition":
		_, ok := model.Transitions[route.Transition]
		return ok
	default:
		return false
	}
}
