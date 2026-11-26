package operator

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
)

var ErrAlreadyBootstrapped = errors.New("operator already bootstrapped")
var ErrInvalidBootstrap = errors.New("operator bootstrap requires username and secret")
var ErrInvalidCredentials = errors.New("invalid operator credentials")
var ErrOperatorRequired = errors.New("operator access required")
var ErrInvalidOperatorAction = errors.New("invalid operator action")

type ActorKind string

const (
	ActorUser     ActorKind = "user"
	ActorOperator ActorKind = "operator"
)

type Actor struct {
	ID         string
	Kind       ActorKind
	sessionKey string
}

type AuditEvent struct {
	Action  string
	ActorID string
	Success bool
	Detail  string
}

type Account struct {
	Username       string
	SecretDigest   string
	MFASecret      string
	RecoveryDigest string
}

func (s *Store) CreateRecoveryToken(tokenSource func() string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.account == nil || s.authenticatedOperator == "" || tokenSource == nil {
		return "", ErrInvalidCredentials
	}
	token := tokenSource()
	if token == "" {
		return "", ErrInvalidCredentials
	}
	digest := sha256.Sum256([]byte(token))
	s.account.RecoveryDigest = hex.EncodeToString(digest[:])
	return token, nil
}

func (s *Store) Recover(token, newSecret string, digest func(string) string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.account == nil || token == "" || newSecret == "" || digest == nil || s.account.RecoveryDigest == "" {
		return ErrInvalidCredentials
	}
	tokenSum := sha256.Sum256([]byte(token))
	actual := hex.EncodeToString(tokenSum[:])
	if actual != s.account.RecoveryDigest {
		return ErrInvalidCredentials
	}
	s.account.SecretDigest = digest(newSecret)
	s.account.RecoveryDigest = ""
	s.account.MFASecret = ""
	s.authenticatedOperator = ""
	s.sessions = nil
	s.recordLocked("operator.recovery", s.account.Username, true, s.account.Username)
	return nil
}

type Store struct {
	mu                    sync.RWMutex
	account               *Account
	audit                 []AuditEvent
	authenticatedOperator string
	sessions              map[string]string
}

func (s *Store) Audit() []AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]AuditEvent(nil), s.audit...)
}

func (s *Store) record(action, actorID string, success bool, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordLocked(action, actorID, success, detail)
}

func (s *Store) recordLocked(action, actorID string, success bool, detail string) {
	s.audit = append(s.audit, AuditEvent{Action: action, ActorID: actorID, Success: success, Detail: detail})
}

func (s *Store) AuthorizeAction(actor Actor, action string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if actor.ID == "" || action == "" || actor.Kind != ActorOperator || s.account == nil || actor.ID != s.account.Username {
		return ErrOperatorRequired
	}
	if actor.sessionKey != "" {
		if s.sessions[actor.sessionKey] != actor.ID {
			return ErrOperatorRequired
		}
	} else if actor.ID != s.authenticatedOperator {
		return ErrOperatorRequired
	}
	return nil
}

func (s *Store) ExecuteAction(actor Actor, action string, operation func() error) error {
	if operation == nil {
		return ErrInvalidOperatorAction
	}
	if err := s.AuthorizeAction(actor, action); err != nil {
		s.record(action, actor.ID, false, err.Error())
		return err
	}
	if err := operation(); err != nil {
		s.record(action, actor.ID, false, err.Error())
		return err
	}
	s.record(action, actor.ID, true, "")
	return nil
}

func (s *Store) Bootstrap(username, secret string, digest func(string) string) (Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.account != nil {
		return Account{}, ErrAlreadyBootstrapped
	}
	if strings.TrimSpace(username) == "" || secret == "" || digest == nil {
		return Account{}, ErrInvalidBootstrap
	}
	a := Account{Username: username, SecretDigest: digest(secret)}
	s.account = &a
	return a, nil
}

func (s *Store) Account() (Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.account == nil {
		return Account{}, false
	}
	return *s.account, true
}

// Authenticate supports trusted in-process operator setup. HTTP callers must
// use AuthenticateSession instead, so independent logins cannot alter sessions.
func (s *Store) Authenticate(username, secret string, digest func(string) string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.authenticateLocked(username, secret, digest); err != nil {
		s.authenticatedOperator = ""
		return err
	}
	s.authenticatedOperator = username
	return nil
}

func (s *Store) authenticateLocked(username, secret string, digest func(string) string) error {
	valid := s.account != nil && digest != nil && username == s.account.Username
	if valid {
		valid = subtle.ConstantTimeCompare([]byte(digest(secret)), []byte(s.account.SecretDigest)) == 1
	}
	s.recordLocked("operator.signin", username, valid, "")
	if !valid {
		return ErrInvalidCredentials
	}
	return nil
}
