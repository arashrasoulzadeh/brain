package operator

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

var ErrInvalidSession = errors.New("invalid operator session")

// DefaultSessionLifetime is the absolute lifetime of an operator bearer token.
const DefaultSessionLifetime = 8 * time.Hour

type operatorSession struct {
	operatorID string
	expiresAt  time.Time
}

// NewStoreWithClock constructs an operator store with a controlled clock.
// A zero-value Store uses the system clock.
func NewStoreWithClock(clock func() time.Time) *Store {
	return &Store{clock: clock}
}

func (s *Store) now() time.Time {
	if s.clock != nil {
		return s.clock()
	}
	return time.Now()
}

// AuthenticateSession verifies credentials and issues a session in one operation.
// It never changes authentication state for other sessions.
func (s *Store) AuthenticateSession(username, secret string, digest func(string) string, tokenSource func() string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.authenticateLocked(username, secret, digest); err != nil {
		return "", err
	}
	return s.issueSessionLocked(username, tokenSource)
}

// IssueSession supports trusted in-process callers using Authenticate.
func (s *Store) IssueSession(tokenSource func() string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.account == nil || s.authenticatedOperator == "" {
		return "", ErrInvalidCredentials
	}
	return s.issueSessionLocked(s.authenticatedOperator, tokenSource)
}

func (s *Store) issueSessionLocked(username string, tokenSource func() string) (string, error) {
	if tokenSource == nil {
		return "", ErrInvalidCredentials
	}
	token := tokenSource()
	if token == "" {
		return "", ErrInvalidSession
	}
	now := s.now()
	for key, existing := range s.sessions {
		if !existing.expiresAt.After(now) {
			delete(s.sessions, key)
		}
	}
	key := sessionKey(token)
	if _, exists := s.sessions[key]; exists {
		return "", ErrInvalidSession
	}
	if s.sessions == nil {
		s.sessions = make(map[string]operatorSession)
	}
	s.sessions[key] = operatorSession{operatorID: username, expiresAt: now.Add(DefaultSessionLifetime)}
	return token, nil
}

func (s *Store) ValidateSession(token string) (Actor, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if token == "" || s.account == nil {
		return Actor{}, false
	}
	key := sessionKey(token)
	if !s.sessionValidLocked(key, s.account.Username) {
		return Actor{}, false
	}
	return Actor{ID: s.account.Username, Kind: ActorOperator, sessionKey: key}, true
}

func (s *Store) sessionValidLocked(key, operatorID string) bool {
	session, ok := s.sessions[key]
	return ok && operatorID != "" && session.operatorID == operatorID && session.expiresAt.After(s.now())
}

func (s *Store) RevokeSession(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if token == "" {
		return ErrInvalidSession
	}
	key := sessionKey(token)
	if _, ok := s.sessions[key]; !ok {
		return ErrInvalidSession
	}
	delete(s.sessions, key)
	return nil
}

func sessionKey(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}
