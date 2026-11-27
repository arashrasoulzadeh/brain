package plugin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"
)

var ErrWebSocketTicketDenied = errors.New("connection ticket denied")

const MaxWebSocketTicketLifetime = time.Minute

// WebSocketTicketStore is a bounded, process-local transport credential store.
// Only a trusted adapter may issue tickets after worker membership admission.
// The required authorizer must recheck the project, actor, resource, grants
// and worker version on issue and redemption; a nil authorizer denies all.
type WebSocketTicketStore struct {
	mu        sync.Mutex
	capacity  int
	authorize func(HostCall, time.Time) error
	tickets   map[[sha256.Size]byte]webSocketTicketBinding
}

func NewWebSocketTicketStore(capacity int, authorize func(HostCall, time.Time) error) *WebSocketTicketStore {
	return &WebSocketTicketStore{
		capacity: capacity, authorize: authorize,
		tickets: make(map[[sha256.Size]byte]webSocketTicketBinding),
	}
}

func validWebSocketTicketCall(call HostCall) bool {
	return isValidProjectID(call.ProjectID) && validAccessIdentity(call.ActorID) &&
		strings.HasPrefix(call.ActorID, call.ProjectID+":") && call.ActorID != call.ProjectID+":" &&
		validAccessIdentity(call.PluginID) && validAccessIdentity(call.Version) &&
		validAccessIdentity(call.Operation) && validAccessIdentity(call.Capability) &&
		validAccessIdentity(call.Resource) && call.Resource != "*"
}

type webSocketTicketBinding struct {
	call    HostCall
	session string
	path    string
	origin  string
}

func (s *WebSocketTicketStore) Issue(call HostCall, now time.Time, lifetime time.Duration) (string, error) {
	return s.issue(webSocketTicketBinding{call: call}, now, lifetime)
}

func (s *WebSocketTicketStore) issue(binding webSocketTicketBinding, now time.Time, lifetime time.Duration) (string, error) {
	call := binding.call
	if s == nil || s.authorize == nil || s.capacity <= 0 ||
		now.IsZero() || lifetime <= 0 || lifetime > MaxWebSocketTicketLifetime || !validWebSocketTicketCall(call) {
		return "", ErrWebSocketTicketDenied
	}
	call.ExpiresAt = now.Add(lifetime)
	if err := s.authorize(call, now); err != nil {
		return "", ErrWebSocketTicketDenied
	}
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", ErrWebSocketTicketDenied
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	digest := sha256.Sum256([]byte(token))
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, existing := range s.tickets {
		if !existing.call.ExpiresAt.After(now) {
			delete(s.tickets, key)
		}
	}
	if len(s.tickets) >= s.capacity {
		return "", ErrWebSocketTicketDenied
	}
	if _, exists := s.tickets[digest]; exists {
		return "", ErrWebSocketTicketDenied
	}
	binding.call = call
	s.tickets[digest] = binding
	return token, nil
}

// Consume atomically redeems a ticket for its original project. The caller
// receives only the trusted binding and must use that binding, never socket
// input, to select actor, worker and resource. It must reauthorize every
// subsequent invocation. A revoked or expired ticket cannot be retried.
func (s *WebSocketTicketStore) Consume(projectID, token string, now time.Time) (HostCall, error) {
	binding, err := s.consume(projectID, token, now)
	return binding.call, err
}

func (s *WebSocketTicketStore) consume(projectID, token string, now time.Time) (webSocketTicketBinding, error) {
	if s == nil || s.authorize == nil || !isValidProjectID(projectID) ||
		len(token) != base64.RawURLEncoding.EncodedLen(32) || now.IsZero() {
		return webSocketTicketBinding{}, ErrWebSocketTicketDenied
	}
	digest := sha256.Sum256([]byte(token))
	s.mu.Lock()
	binding, found := s.tickets[digest]
	call := binding.call
	if !found || call.ProjectID != projectID {
		s.mu.Unlock()
		return webSocketTicketBinding{}, ErrWebSocketTicketDenied
	}
	delete(s.tickets, digest)
	s.mu.Unlock()
	if !call.ExpiresAt.After(now) || s.authorize(call, now) != nil {
		return webSocketTicketBinding{}, ErrWebSocketTicketDenied
	}
	return binding, nil
}

// RevokeProject invalidates outstanding tickets on project shutdown. Existing
// connections must separately be closed through WebSocketRegistry.
func (s *WebSocketTicketStore) RevokeProject(projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for digest, binding := range s.tickets {
		if binding.call.ProjectID == projectID {
			delete(s.tickets, digest)
		}
	}
}
