package deploy

import "errors"

var ErrIncompatibleRollbackSchema = errors.New("rollback schema is incompatible")

func SelectRollbackManifest(manifest ReleaseManifest, schemaCompatible bool) (ReleaseManifest, error) {
	if !schemaCompatible {
		return ReleaseManifest{}, ErrIncompatibleRollbackSchema
	}
	if err := ValidateReleaseManifest(manifest); err != nil {
		return ReleaseManifest{}, err
	}
	return manifest, nil
}
