package plugin

import (
	"errors"
	"math"
	"strings"
)

var ErrManifestStorageUnavailable = errors.New("manifest storage unavailable")

// StorageSpec selects a bounded storage primitive. Namespaces and project
// identity are assigned by the host, never by the manifest or request payload.
type StorageSpec struct {
	Action   string         `yaml:"action" json:"action"`
	Key      string         `yaml:"key,omitempty" json:"key,omitempty"`
	FixedKey string         `yaml:"fixed_key,omitempty" json:"fixed_key,omitempty"`
	Target   string         `yaml:"target,omitempty" json:"target,omitempty"`
	Field    string         `yaml:"field,omitempty" json:"field,omitempty"`
	Input    string         `yaml:"input,omitempty" json:"input,omitempty"`
	Values   map[string]any `yaml:"values,omitempty" json:"values,omitempty"`
}

func (s *StorageSpec) valid() bool {
	if s == nil {
		return true
	}
	if (s.Key == "") == (s.FixedKey == "") || (s.Key != "" && !validResponseKey(s.Key)) || (s.FixedKey != "" && !isValidProjectID(s.FixedKey)) || len(s.Values) > 32 {
		return false
	}
	if s.Target != "" && (s.Target == s.Key || s.Target == s.Input) || s.Action == "blob.read" && s.Field == s.Key {
		return false
	}
	for _, name := range []string{s.Target, s.Field, s.Input} {
		if name != "" && !validResponseKey(name) {
			return false
		}
	}
	for key, value := range s.Values {
		if !validResponseKey(key) {
			return false
		}
		switch value := value.(type) {
		case string:
			if strings.HasPrefix(value, "$") && (!strings.HasPrefix(value, "$input.") || !validResponseKey(strings.TrimPrefix(value, "$input."))) {
				return false
			}
		case []any:
			if len(value) != 0 {
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
	switch s.Action {
	case "blob.put", "rows.append":
		return s.Input != "" && s.Target == "" && s.Field == "" && len(s.Values) == 0
	case "blob.read":
		return s.Target != "" && s.Field != "" && s.Target != s.Field && s.Input == "" && len(s.Values) == 0
	case "blob.publish", "rows.read", "state.read":
		return s.Target != "" && s.Field == "" && s.Input == "" && len(s.Values) == 0
	case "state.create":
		return s.Target != "" && len(s.Values) > 0 && s.Field == "" && s.Input == ""
	case "state.patch":
		return s.Target != "" && ((len(s.Values) > 0 && s.Field == "" && s.Input == "") || (len(s.Values) == 0 && s.Field != "" && s.Input != ""))
	case "state.add_unique":
		return s.Target != "" && s.Field != "" && s.Input != "" && len(s.Values) == 0
	case "state.append":
		return s.Field != "" && s.Target == "" && s.Input == "" && len(s.Values) == 0
	default:
		return false
	}
}
