package operator

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
)

var ErrAlreadyBootstrapped = errors.New("operator already bootstrapped")
var ErrInvalidBootstrap = errors.New("operator bootstrap requires username and secret")
var ErrInvalidCredentials = errors.New("invalid operator credentials")

type Account struct {
	Username       string
	SecretDigest   string
	MFASecret      string
	RecoveryDigest string
}

func (s *Store) CreateRecoveryToken(tokenSource func() string) (string, error) {
	if s.account == nil || tokenSource == nil {
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
	if s.account == nil || token == "" || newSecret == "" || s.account.RecoveryDigest == "" {
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
	return nil
}

type Store struct{ account *Account }

func (s *Store) Bootstrap(username, secret string, digest func(string) string) (Account, error) {
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
	if s.account == nil {
		return Account{}, false
	}
	return *s.account, true
}

func (s *Store) Authenticate(username, secret string, digest func(string) string) error {
	if s.account == nil || digest == nil || username != s.account.Username {
		return ErrInvalidCredentials
	}
	got := digest(secret)
	if subtle.ConstantTimeCompare([]byte(got), []byte(s.account.SecretDigest)) != 1 {
		return ErrInvalidCredentials
	}
	return nil
}
