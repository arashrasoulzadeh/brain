package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestValidateReleaseSHARequiresFullLowercaseCommit(t *testing.T) {
	if err := deploy.ValidateReleaseSHA("0123456789abcdef0123456789abcdef01234567"); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"abc", "ABCDEF0123456789abcdef0123456789abcdef01", "0123456789abcdef0123456789abcdef012345678"} {
		if !errors.Is(deploy.ValidateReleaseSHA(value), deploy.ErrInvalidReleaseSHA) {
			t.Fatalf("invalid release SHA accepted: %q", value)
		}
	}
}
