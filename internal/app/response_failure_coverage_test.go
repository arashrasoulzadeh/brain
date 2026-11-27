package app

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brain/internal/operator"
)

type coverageUnavailableEntropy struct{}

func (coverageUnavailableEntropy) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestCoverageNormalizeEmptyAPIError(t *testing.T) {
	var message APIMessage
	if err := json.Unmarshal(normalizeAPIError(http.StatusForbidden, []byte(" \n\t ")), &message); err != nil {
		t.Fatal(err)
	}
	if message.Status != "error" || message.Code != "Forbidden" || message.Message != "Forbidden" {
		t.Fatalf("unexpected error message: %+v", message)
	}
}

func TestCoverageOperatorLoginEntropyFailure(t *testing.T) {
	store := &operator.Store{}
	digest := func(value string) string { return value }
	if _, err := store.Bootstrap("operator", "secret", digest); err != nil {
		t.Fatal(err)
	}
	original := rand.Reader
	rand.Reader = coverageUnavailableEntropy{}
	t.Cleanup(func() { rand.Reader = original })
	request := httptest.NewRequest(http.MethodPost, "/admin/session", strings.NewReader(`{"username":"operator","secret":"secret"}`))
	response := httptest.NewRecorder()
	OperatorSessionHandler(store, digest).ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), "operator session unavailable") {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
}
