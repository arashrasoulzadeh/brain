package deploy

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

var ErrInvalidObjectVersionManifest = errors.New("invalid object version manifest")

type ObjectVersion struct {
	ProjectID string `json:"project_id"`
	ObjectKey string `json:"object_key"`
	VersionID string `json:"version_id"`
}

type ObjectVersionManifest struct {
	Objects []ObjectVersion `json:"objects"`
}

func MissingObjectVersions(objects []ObjectVersion, available map[string]bool) []ObjectVersion {
	missing := make([]ObjectVersion, 0)
	for _, object := range objects {
		identity := object.ProjectID + "\x00" + object.ObjectKey + "\x00" + object.VersionID
		if !available[identity] {
			missing = append(missing, object)
		}
	}
	return missing
}

func ValidateObjectVersionManifest(manifest ObjectVersionManifest) error {
	seen := make(map[string]struct{}, len(manifest.Objects))
	for _, object := range manifest.Objects {
		if strings.TrimSpace(object.ProjectID) == "" || strings.TrimSpace(object.ObjectKey) == "" || strings.TrimSpace(object.VersionID) == "" {
			return ErrInvalidObjectVersionManifest
		}
		identity := object.ProjectID + "\x00" + object.ObjectKey + "\x00" + object.VersionID
		if _, exists := seen[identity]; exists {
			return ErrInvalidObjectVersionManifest
		}
		seen[identity] = struct{}{}
	}
	return nil
}

func RecordObjectVersionManifest(path string, manifest ObjectVersionManifest) error {
	if err := ValidateObjectVersionManifest(manifest); err != nil {
		return err
	}
	return recordJSON(path, manifest)
}

func ReadObjectVersionManifest(path string) (ObjectVersionManifest, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return ObjectVersionManifest{}, err
	}
	var manifest ObjectVersionManifest
	if err := json.Unmarshal(contents, &manifest); err != nil {
		return ObjectVersionManifest{}, ErrInvalidObjectVersionManifest
	}
	if err := ValidateObjectVersionManifest(manifest); err != nil {
		return ObjectVersionManifest{}, err
	}
	return manifest, nil
}
