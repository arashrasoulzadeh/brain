package plugin

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var ErrActorDenied = errors.New("actor permission denied")
var ErrInvocationExpired = errors.New("invocation credential expired")
var ErrWorkerUnavailable = errors.New("plugin worker unavailable")
var ErrPluginScopeDenied = errors.New("plugin scope denied")
var ErrRegistryUnavailable = errors.New("plugin registry unavailable")

type HostCall struct {
	ProjectID        string
	ActorID          string
	PluginID         string
	Version          string
	Capability       string
	Operation        string
	Resource         string
	MatchmakingScope string
	ExpiresAt        time.Time
}

type Broker struct {
	Registry          *Registry
	ActorAllows       func(actorID, capability, resource string) bool
	WorkerAvailable   func(pluginID, version string) bool
	MatchmakingAllows func(projectID, scope string) bool
	issuedMu          sync.Mutex
	issuedTokens      map[string]HostCall
}

type InvocationCredential struct {
	Token string
	Call  HostCall
}

func (b *Broker) Issue(call HostCall, now time.Time, lifetime time.Duration) (InvocationCredential, error) {
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
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	b.issuedMu.Lock()
	defer b.issuedMu.Unlock()
	if b.issuedTokens == nil {
		b.issuedTokens = map[string]HostCall{}
	}
	b.issuedTokens[token] = call
	return InvocationCredential{Token: token, Call: call}, nil
}

func (b *Broker) Authorize(call HostCall, now time.Time) error {
	if b == nil {
		return ErrRegistryUnavailable
	}
	if !isValidProjectID(call.ProjectID) || !validAccessIdentity(call.ActorID) {
		return ErrProjectInvalid
	}
	if !call.ExpiresAt.After(now) {
		return ErrInvocationExpired
	}
	if b.WorkerAvailable != nil && !b.WorkerAvailable(call.PluginID, call.Version) {
		return ErrWorkerUnavailable
	}
	if call.Operation != "" {
		if b.Registry == nil {
			return ErrRegistryUnavailable
		}
		b.Registry.readinessMu.RLock()
		registered, ok := b.Registry.registered[call.PluginID+"@"+call.Version]
		b.Registry.readinessMu.RUnlock()
		if !ok {
			return ErrNotRegistered
		}
		route, ok := registered.Manifest.Route(call.Operation)
		if !ok || route.Capability != call.Capability {
			return ErrActorDenied
		}
	}
	if call.MatchmakingScope != "" && (b.MatchmakingAllows == nil || !b.MatchmakingAllows(call.ProjectID, call.MatchmakingScope)) {
		return ErrPluginScopeDenied
	}
	if b.ActorAllows == nil || !b.ActorAllows(call.ActorID, call.Capability, call.Resource) {
		return ErrActorDenied
	}
	if b.Registry == nil {
		return ErrRegistryUnavailable
	}
	if err := b.Registry.authorizeActorAccess(call); err != nil {
		return err
	}
	if call.Resource != "" {
		return b.Registry.AuthorizeResource(call.ProjectID, call.PluginID, call.Version, call.Capability, call.Resource)
	}
	return b.Registry.Authorize(call.ProjectID, call.PluginID, call.Version, call.Capability)
}

// Commit authorizes a raw call for low-level broker tests. Production code must use CommitCredential.
func (b *Broker) Commit(call HostCall, now time.Time) error {
	return b.Authorize(call, now)
}

func (b *Broker) CommitCredential(credential InvocationCredential, now time.Time) error {
	b.issuedMu.Lock()
	defer b.issuedMu.Unlock()
	if credential.Token == "" || b.issuedTokens == nil {
		return ErrInvocationExpired
	}
	issued, ok := b.issuedTokens[credential.Token]
	if !ok || issued != credential.Call {
		return ErrInvocationExpired
	}
	delete(b.issuedTokens, credential.Token)
	return b.Authorize(issued, now)
}
