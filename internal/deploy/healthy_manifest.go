package deploy

import (
	"encoding/json"
	"os"
)

func RecordHealthyManifest(path string, manifest ReleaseManifest) error {
	if err := ValidateReleaseManifest(manifest); err != nil {
		return err
	}
	return recordJSON(path, manifest)
}

func ReadHealthyManifest(path string) (ReleaseManifest, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return ReleaseManifest{}, err
	}
	var manifest ReleaseManifest
	if err := json.Unmarshal(contents, &manifest); err != nil {
		return ReleaseManifest{}, ErrInvalidReleaseManifest
	}
	if err := ValidateReleaseManifest(manifest); err != nil {
		return ReleaseManifest{}, err
	}
	return manifest, nil
}
