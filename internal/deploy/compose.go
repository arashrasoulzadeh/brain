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
	if len(document.Services["brain"]) == 0 || len(document.Services["postgres"]) == 0 {
		return ErrInvalidComposeConfig
	}
	for _, service := range []map[string]any{document.Services["brain"], document.Services["postgres"]} {
		if serviceUsesHostIsolation(service) {
			return ErrInvalidComposeConfig
		}
	}
	return nil
}

func serviceUsesHostIsolation(service map[string]any) bool {
	if privileged, ok := service["privileged"].(bool); ok && privileged {
		return true
	}
	if networkMode, ok := service["network_mode"].(string); ok && networkMode == "host" {
		return true
	}
	if pidMode, ok := service["pid"].(string); ok && pidMode == "host" {
		return true
	}
	return false
}
