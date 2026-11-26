package operator

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrInvalidSession = errors.New("invalid operator session")

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
	key := sessionKey(token)
	if _, exists := s.sessions[key]; exists {
		return "", ErrInvalidSession
	}
	if s.sessions == nil {
		s.sessions = make(map[string]string)
	}
	s.sessions[key] = username
	return token, nil
}

func (s *Store) ValidateSession(token string) (Actor, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if token == "" || s.account == nil {
		return Actor{}, false
	}
	key := sessionKey(token)
	operatorID, ok := s.sessions[key]
	if !ok || operatorID == "" || operatorID != s.account.Username {
		return Actor{}, false
	}
	return Actor{ID: operatorID, Kind: ActorOperator, sessionKey: key}, true
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
