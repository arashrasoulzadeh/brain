package plugin

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestConfiguredFileUploadLimit(t *testing.T) {
	for _, value := range []any{1, float64(1024), adminFileUploadLimit} {
		if limit, ok := configuredFileUploadLimit(map[string]any{"maxUploadBytes": value}); !ok || limit < 1 {
			t.Fatalf("valid limit %v rejected", value)
		}
	}
	for _, value := range []any{nil, "1024", 0, -1, 1.5, adminFileUploadLimit + 1} {
		if _, ok := configuredFileUploadLimit(map[string]any{"maxUploadBytes": value}); ok {
			t.Fatalf("invalid limit %v accepted", value)
		}
	}
	config := map[string]any{"maxUploadBytes": 3}
	for _, tc := range []struct {
		data string
		want bool
	}{
		{"abc", false},
		{"abcd", true},
		{"éé", true},
		{adminFileDataPrefix + "AP+A", false},
		{adminFileDataPrefix + "AP+AAQ==", true},
		{adminFileDataPrefix + "invalid", true},
	} {
		payload, err := json.Marshal(map[string]string{"key": "fixture", "data": tc.data})
		if err != nil {
			t.Fatal(err)
		}
		request := WorkerRequest{PluginID: "files", Operation: "files.upload", Payload: payload}
		err = validateConfiguredFileUpload(request, config)
		if errors.Is(err, ErrInvalidWorkerRequest) != tc.want {
			t.Fatalf("data %q: error = %v, want rejection %v", tc.data, err, tc.want)
		}
		if err := validateConfiguredFileUpload(request, nil); err != nil {
			t.Fatal("unconfigured legacy request was restricted", err)
		}
		request.Operation = "files.download"
		if err := validateConfiguredFileUpload(request, config); err != nil {
			t.Fatal("download was restricted", err)
		}
	}
	request := WorkerRequest{PluginID: "files", Operation: "files.upload", Payload: []byte(strings.Repeat("x", 5))}
	if err := validateConfiguredFileUpload(request, config); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("invalid JSON accepted: %v", err)
	}
}

func TestFileUploadConfigurationIsProjectScoped(t *testing.T) {
	r := NewRegistry(1)
	m := Manifest{ID: "files", Version: "1", HostAPI: 1, ConfigSchema: map[string]string{"maxUploadBytes": "int", "bucket": "string"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"one", "two"} {
		if err := r.Install(id, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Configure("one", m.ID, m.Version, map[string]any{"maxUploadBytes": 1024, "bucket": "private-value"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure("one", m.ID, m.Version, map[string]any{"maxUploadBytes": adminFileUploadLimit + 1}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("oversized setting accepted: %v", err)
	}
	one, err := r.Status("one", m.ID, m.Version)
	if err != nil || one.MaxUploadBytes != 1024 || len(one.ConfiguredFields) != 2 {
		t.Fatalf("wrong first project status: %#v %v", one, err)
	}
	two, err := r.Status("two", m.ID, m.Version)
	if err != nil || two.MaxUploadBytes != adminFileUploadLimit || len(two.ConfiguredFields) != 0 {
		t.Fatalf("wrong second project status: %#v %v", two, err)
	}
	serialized, err := json.Marshal(one)
	if err != nil || strings.Contains(string(serialized), "private-value") {
		t.Fatal("configuration secret leaked")
	}
}
