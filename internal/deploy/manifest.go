package deploy

import "errors"

var ErrInvalidReleaseManifest = errors.New("invalid release manifest")

type ReleaseManifest struct {
	CommitSHA                  string `json:"commit_sha"`
	Backend                    string `json:"backend"`
	Compiler                   string `json:"compiler"`
	Runner                     string `json:"runner"`
	PackagedIsolationVerified  bool   `json:"packaged_isolation_verified"`
	PackagedRevocationVerified bool   `json:"packaged_revocation_verified"`
}

func ValidateReleaseManifest(manifest ReleaseManifest) error {
	if ValidateReleaseSHA(manifest.CommitSHA) != nil || manifest.Backend == "" ||
		manifest.Compiler == "" || manifest.Runner == "" || !manifest.PackagedIsolationVerified || !manifest.PackagedRevocationVerified {
		return ErrInvalidReleaseManifest
	}
	return nil
}
