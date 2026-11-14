package deploy

import (
	"encoding/json"
	"os"
)

func RecordHealthyManifest(path string, manifest ReleaseManifest) error {
	if err := ValidateReleaseManifest(manifest); err != nil {
		return err
	}
	contents, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	temporaryPath := path + ".tmp"
	if err := os.WriteFile(temporaryPath, append(contents, '\n'), 0o600); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	return nil
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
