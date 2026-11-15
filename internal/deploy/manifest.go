package deploy

import (
	"errors"
	"regexp"
)

var ErrInvalidReleaseManifest = errors.New("invalid release manifest")
var digestPinnedImage = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*@sha256:[0-9a-f]{64}$`)

type ReleaseManifest struct {
	CommitSHA                  string `json:"commit_sha"`
	Backend                    string `json:"backend"`
	Compiler                   string `json:"compiler"`
	Runner                     string `json:"runner"`
	PackagedIsolationVerified  bool   `json:"packaged_isolation_verified"`
	PackagedRevocationVerified bool   `json:"packaged_revocation_verified"`
}

func ValidateReleaseManifest(manifest ReleaseManifest) error {
	if ValidateReleaseSHA(manifest.CommitSHA) != nil ||
		!digestPinnedImage.MatchString(manifest.Backend) ||
		!digestPinnedImage.MatchString(manifest.Compiler) ||
		!digestPinnedImage.MatchString(manifest.Runner) ||
		!manifest.PackagedIsolationVerified || !manifest.PackagedRevocationVerified {
		return ErrInvalidReleaseManifest
	}
	return nil
}
