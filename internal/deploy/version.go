package deploy

import (
	"os"
	"strings"
)

func RecordDeployedVersion(path string, releaseSHA string) error {
	if err := ValidateReleaseSHA(releaseSHA); err != nil {
		return err
	}
	return writeAtomicRecord(path, []byte(releaseSHA+"\n"))
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
