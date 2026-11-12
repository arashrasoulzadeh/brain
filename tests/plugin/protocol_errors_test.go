package plugin_test

import (
	"testing"

	"brain/internal/plugin"
)

func TestProtocolErrorUsesBoundedCodesAndRequestID(t *testing.T) {
	error := plugin.ProtocolError{RequestID: "req-1", Code: plugin.ErrorConflict, Message: "room changed"}
	if !error.Valid() {
		t.Fatal("valid protocol error rejected")
	}
	error.Code = "secret_detail"
	if error.Valid() {
		t.Fatal("unbounded error code accepted")
	}
}
