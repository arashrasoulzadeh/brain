package deploy

import (
	"errors"
	"regexp"
)

var ErrInvalidReleaseSHA = errors.New("invalid release sha")
var releaseSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

func ValidateReleaseSHA(value string) error {
	if !releaseSHA.MatchString(value) {
		return ErrInvalidReleaseSHA
	}
	return nil
}
