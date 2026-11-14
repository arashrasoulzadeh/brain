package deploy

import "errors"

var ErrInvalidReleaseManifest = errors.New("invalid release manifest")

type ReleaseManifest struct {
	CommitSHA string `json:"commit_sha"`
	Backend   string `json:"backend"`
	Compiler  string `json:"compiler"`
	Runner    string `json:"runner"`
}

func ValidateReleaseManifest(manifest ReleaseManifest) error {
	if ValidateReleaseSHA(manifest.CommitSHA) != nil || manifest.Backend == "" ||
		manifest.Compiler == "" || manifest.Runner == "" {
		return ErrInvalidReleaseManifest
	}
	return nil
}
