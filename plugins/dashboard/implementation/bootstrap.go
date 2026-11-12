package dashboard

import "brain/internal/plugin"

func Bootstrap(hostAPI int) (*plugin.Registry, error) {
	registry := plugin.NewRegistry(hostAPI)
	if err := registry.Register(Manifest()); err != nil {
		return nil, err
	}
	if err := registry.SetReadiness("dashboard", "0.1.0", true); err != nil {
		return nil, err
	}
	return registry, nil
}
