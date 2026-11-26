package plugin

import (
	"math"
	"strings"
)

// ResponseSpec describes bounded input validation and response mapping, not executable code.
type ResponseSpec struct {
	Input  []ResponseField `yaml:"input" json:"input"`
	Output map[string]any  `yaml:"output" json:"output"`
}

type ResponseField struct {
	Name     string  `yaml:"name" json:"name"`
	Type     string  `yaml:"type" json:"type"`
	Trim     bool    `yaml:"trim,omitempty" json:"trim,omitempty"`
	Required bool    `yaml:"required,omitempty" json:"required,omitempty"`
	Default  *string `yaml:"default,omitempty" json:"default,omitempty"`
	Error    string  `yaml:"error,omitempty" json:"error,omitempty"`
}

func validResponseKey(key string) bool {
	if key == "" || key == "__proto__" || key == "constructor" || key == "prototype" {
		return false
	}
	for i, c := range key {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || i > 0 && c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

func (s *ResponseSpec) valid() bool {
	if s == nil || len(s.Input) > 32 || len(s.Output) == 0 || len(s.Output) > 32 {
		return false
	}
	fields := map[string]bool{}
	for _, field := range s.Input {
		if !validResponseKey(field.Name) || fields[field.Name] || (field.Type != "string" && field.Type != "value") ||
			(field.Type != "string" && (field.Trim || field.Default != nil)) || (field.Required && field.Error == "") {
			return false
		}
		fields[field.Name] = true
	}
	for key, value := range s.Output {
		if !validResponseKey(key) {
			return false
		}
		switch value := value.(type) {
		case string:
			if strings.HasPrefix(value, "$") && value != "$project.id" && value != "$plugin.id" && (!strings.HasPrefix(value, "$input.") || !fields[strings.TrimPrefix(value, "$input.")]) {
				return false
			}
		case float64:
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return false
			}
		case bool, int, int64, uint64, nil:
		default:
			return false
		}
	}
	return true
}
