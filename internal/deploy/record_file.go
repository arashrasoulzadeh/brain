package deploy

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func recordJSON(path string, value any) error {
	contents, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return writeAtomicRecord(path, append(contents, '\n'))
}

func writeAtomicRecord(path string, contents []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".brain-record-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), path)
}
