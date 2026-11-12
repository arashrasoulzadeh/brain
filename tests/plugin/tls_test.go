package plugin_test

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"brain/internal/plugin"
)

func TestExternalTLSRequiresMutualTLS13(t *testing.T) {
	pool := x509.NewCertPool()
	config := plugin.ExternalTLSConfig(tls.Certificate{}, pool)
	if config.MinVersion != tls.VersionTLS13 || config.ClientAuth != tls.RequireAndVerifyClientCert || config.ClientCAs != pool {
		t.Fatalf("TLS config = %#v", config)
	}
}
