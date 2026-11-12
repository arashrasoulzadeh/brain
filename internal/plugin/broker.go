package plugin

import (
	"errors"
	"time"
)

var ErrActorDenied = errors.New("actor permission denied")
var ErrInvocationExpired = errors.New("invocation credential expired")

type HostCall struct {
	ProjectID  string
	ActorID    string
	PluginID   string
	Version    string
	Capability string
	Resource   string
	ExpiresAt  time.Time
}

type Broker struct {
	Registry    *Registry
	ActorAllows func(actorID, capability, resource string) bool
}

func (b Broker) Authorize(call HostCall, now time.Time) error {
	if call.ProjectID == "" || call.ActorID == "" {
		return ErrProjectInvalid
	}
	if !call.ExpiresAt.After(now) {
		return ErrInvocationExpired
	}
	if b.ActorAllows == nil || !b.ActorAllows(call.ActorID, call.Capability, call.Resource) {
		return ErrActorDenied
	}
	if call.Resource != "" {
		return b.Registry.AuthorizeResource(call.ProjectID, call.PluginID, call.Version, call.Capability, call.Resource)
	}
	return b.Registry.Authorize(call.ProjectID, call.PluginID, call.Version, call.Capability)
}
