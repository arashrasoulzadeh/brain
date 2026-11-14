package deploy

import "errors"

var ErrInvalidReleaseManifest = errors.New("invalid release manifest")

type ReleaseManifest struct {
	CommitSHA string
	Backend   string
	Compiler  string
	Runner    string
}

func ValidateReleaseManifest(manifest ReleaseManifest) error {
	if ValidateReleaseSHA(manifest.CommitSHA) != nil || manifest.Backend == "" ||
		manifest.Compiler == "" || manifest.Runner == "" {
		return ErrInvalidReleaseManifest
	}
	return nil
}
