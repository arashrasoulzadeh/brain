package plugin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// The admin JSON envelope is capped at 64 KiB; leave room for base64 and metadata.
const adminFileUploadLimit = 40 * 1024
const adminFileDataPrefix = "data:application/octet-stream;base64,"

func configuredFileUploadLimit(config map[string]any) (int, bool) {
	value := config["maxUploadBytes"]
	var number float64
	switch v := value.(type) {
	case int:
		number = float64(v)
	case float64:
		number = v
	default:
		return 0, false
	}
	if !(number >= 1 && number <= adminFileUploadLimit) || number != float64(int(number)) {
		return 0, false
	}
	return int(number), true
}

// Explicit per-installation limits also apply to direct API callers; absent
// configuration preserves the historical string storage contract.
func validateConfiguredFileUpload(request WorkerRequest, config map[string]any) error {
	if request.PluginID != "files" || request.Operation != "files.upload" {
		return nil
	}
	limit, ok := configuredFileUploadLimit(config)
	if !ok {
		return nil
	}
	var payload struct {
		Data string `json:"data"`
	}
	if json.Unmarshal(request.Payload, &payload) != nil {
		return ErrInvalidWorkerRequest
	}
	size := len(payload.Data)
	if encoded, found := strings.CutPrefix(payload.Data, adminFileDataPrefix); found {
		if decoded, err := base64.StdEncoding.Strict().DecodeString(encoded); err == nil && base64.StdEncoding.EncodeToString(decoded) == encoded {
			size = len(decoded)
		}
	}
	if size > limit {
		return fmt.Errorf("%w: file exceeds project upload limit (%d bytes)", ErrInvalidWorkerRequest, limit)
	}
	return nil
}
