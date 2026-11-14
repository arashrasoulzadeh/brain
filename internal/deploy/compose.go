package deploy

import (
	"errors"

	"gopkg.in/yaml.v3"
)

var ErrInvalidComposeConfig = errors.New("invalid compose configuration")

func ValidateComposeConfig(raw []byte) error {
	var document struct {
		Services map[string]map[string]any `yaml:"services"`
	}
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return ErrInvalidComposeConfig
	}
	if document.Services["brain"] == nil || document.Services["postgres"] == nil {
		return ErrInvalidComposeConfig
	}
	return nil
}
