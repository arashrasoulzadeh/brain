package deploy

import (
	"os"
	"strings"
)

func RecordDeployedVersion(path string, releaseSHA string) error {
	if err := ValidateReleaseSHA(releaseSHA); err != nil {
		return err
	}
	temporaryPath := path + ".tmp"
	if err := os.WriteFile(temporaryPath, []byte(releaseSHA+"\n"), 0o600); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return err
	}
	return nil
}

func ReadDeployedVersion(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	releaseSHA := strings.TrimSpace(string(contents))
	if err := ValidateReleaseSHA(releaseSHA); err != nil {
		return "", err
	}
	return releaseSHA, nil
}
