package plugin

import (
	"crypto/rand"
	"encoding/base64"
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

type InvocationCredential struct {
	Token string
	Call  HostCall
}

func (b Broker) Issue(call HostCall, now time.Time, lifetime time.Duration) (InvocationCredential, error) {
	if lifetime <= 0 {
		return InvocationCredential{}, ErrInvocationExpired
	}
	call.ExpiresAt = now.Add(lifetime)
	if err := b.Authorize(call, now); err != nil {
		return InvocationCredential{}, err
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return InvocationCredential{}, err
	}
	return InvocationCredential{Token: base64.RawURLEncoding.EncodeToString(raw[:]), Call: call}, nil
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

func (b Broker) Commit(call HostCall, now time.Time) error {
	return b.Authorize(call, now)
}
