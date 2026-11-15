package operator

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrInvalidSession = errors.New("invalid operator session")

func (s *Store) IssueSession(tokenSource func() string) (string, error) {
	if s.account == nil || s.authenticatedOperator == "" || tokenSource == nil {
		return "", ErrInvalidCredentials
	}
	token := tokenSource()
	if token == "" {
		return "", ErrInvalidSession
	}
	if s.sessions == nil {
		s.sessions = make(map[string]string)
	}
	digest := sha256.Sum256([]byte(token))
	s.sessions[hex.EncodeToString(digest[:])] = s.authenticatedOperator
	return token, nil
}

func (s *Store) ValidateSession(token string) (Actor, bool) {
	if token == "" || s.sessions == nil {
		return Actor{}, false
	}
	digest := sha256.Sum256([]byte(token))
	operatorID, ok := s.sessions[hex.EncodeToString(digest[:])]
	if !ok || operatorID == "" || operatorID != s.authenticatedOperator {
		return Actor{}, false
	}
	return Actor{ID: operatorID, Kind: ActorOperator}, true
}

func (s *Store) RevokeSession(token string) error {
	if token == "" || s.sessions == nil {
		return ErrInvalidSession
	}
	digest := sha256.Sum256([]byte(token))
	key := hex.EncodeToString(digest[:])
	if _, ok := s.sessions[key]; !ok {
		return ErrInvalidSession
	}
	delete(s.sessions, key)
	return nil
}
