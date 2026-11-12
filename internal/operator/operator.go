package operator

import (
	"errors"
	"strings"
)

var ErrAlreadyBootstrapped = errors.New("operator already bootstrapped")
var ErrInvalidBootstrap = errors.New("operator bootstrap requires username and secret")

type Account struct {
	Username     string
	SecretDigest string
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
