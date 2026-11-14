package deploy

import "errors"

var (
	ErrInvalidRestorePair    = errors.New("invalid restore pair")
	ErrIncompleteRestorePair = errors.New("database and object data must be restored together")
)

type RestorePair struct {
	ReleaseSHA             string `json:"release_sha"`
	DatabaseRestored       bool   `json:"database_restored"`
	ObjectManifestRestored bool   `json:"object_manifest_restored"`
}

func (p RestorePair) Validate() error {
	if ValidateReleaseSHA(p.ReleaseSHA) != nil {
		return ErrInvalidRestorePair
	}
	if !p.DatabaseRestored || !p.ObjectManifestRestored {
		return ErrIncompleteRestorePair
	}
	return nil
}
