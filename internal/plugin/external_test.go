package plugin_test

import (
	"errors"
	"testing"

	"brain/internal/plugin"
)

func TestExternalWorkerHandshakeBindsServiceAndProject(t *testing.T) {
	h := plugin.Handshake{Protocol: 1, ProjectID: "dev", PluginID: "arena", Version: "1", ServiceID: "arena-dev"}
	if err := plugin.ValidateHandshake(h, 1, "arena-dev"); err != nil {
		t.Fatal(err)
	}
	h.ProjectID = ""
	if !errors.Is(plugin.ValidateHandshake(h, 1, "arena-dev"), plugin.ErrInvalidHandshake) {
		t.Fatal("unscoped handshake accepted")
	}
}

func TestExternalWorkerHandshakeRejectsWrongIdentity(t *testing.T) {
	h := plugin.Handshake{Protocol: 1, ProjectID: "dev", PluginID: "arena", Version: "1", ServiceID: "other"}
	if !errors.Is(plugin.ValidateHandshake(h, 1, "arena-dev"), plugin.ErrInvalidHandshake) {
		t.Fatal("wrong service accepted")
	}
}

func TestExternalHandshakeAuthenticatesSignature(t *testing.T) {
	h := plugin.Handshake{Protocol: 1, ProjectID: "dev", PluginID: "arena", Version: "1", ServiceID: "arena-dev", Signature: "bad"}
	if plugin.AuthenticateHandshake(h, "secret") {
		t.Fatal("invalid signature accepted")
	}
}
